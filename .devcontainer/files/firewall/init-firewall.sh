#!/bin/bash
# Locks the devcontainer's outbound traffic down to an allowlist of GitHub, Google and a fixed set of domains.
set -euo pipefail  # Exit on error, undefined vars, and pipeline failures
IFS=$'\n\t'       # Stricter word splitting

# Flushes the existing rules and ipset, keeping only the internal Docker DNS resolution
function flushRules() {
  local docker_dns_rules

  # 1. Extract Docker DNS info BEFORE any flushing
  docker_dns_rules=$(iptables-save -t nat | grep "127\.0\.0\.11" || true)

  # Flush existing rules and delete existing ipsets
  iptables -F
  iptables -X
  iptables -t nat -F
  iptables -t nat -X
  iptables -t mangle -F
  iptables -t mangle -X
  ipset destroy allowed-domains 2>/dev/null || true

  # 2. Selectively restore ONLY internal Docker DNS resolution
  if [[ -n "$docker_dns_rules" ]]; then
    echo "Restoring Docker DNS rules..."
    iptables -t nat -N DOCKER_OUTPUT 2>/dev/null || true
    iptables -t nat -N DOCKER_POSTROUTING 2>/dev/null || true
    echo "$docker_dns_rules" | xargs -L 1 iptables -t nat
  else
    echo "No Docker DNS rules to restore"
  fi
}

# Allows DNS, SSH and localhost traffic and creates the empty ipset of allowed destinations
function allowBaseTraffic() {
  # First allow DNS and localhost before any restrictions
  # Allow outbound DNS
  iptables -A OUTPUT -p udp --dport 53 -j ACCEPT
  # Allow inbound DNS responses
  iptables -A INPUT -p udp --sport 53 -j ACCEPT
  # Allow outbound SSH
  iptables -A OUTPUT -p tcp --dport 22 -j ACCEPT
  # Allow inbound SSH responses
  iptables -A INPUT -p tcp --sport 22 -m state --state ESTABLISHED -j ACCEPT
  # Allow localhost
  iptables -A INPUT -i lo -j ACCEPT
  iptables -A OUTPUT -o lo -j ACCEPT

  # Create ipset with CIDR support
  ipset create allowed-domains hash:net
}

# Prints the JSON published at a url, failing when it is empty or lacks what the required jq filter checks
function fetchRanges() {
  local label="$1" url="$2" required_filter="$3" ranges

  ranges=$(curl -s "$url")
  if [[ -z "$ranges" ]]; then
    echo "ERROR: Failed to fetch $label IP ranges" >&2
    exit 1
  fi

  if ! echo "$ranges" | jq -e "$required_filter" >/dev/null; then
    echo "ERROR: $label IP ranges response missing required fields" >&2
    exit 1
  fi

  echo "$ranges"
}

# Aggregates the IPv4 ranges the jq filter selects from the JSON and adds them to the ipset, or excludes them from it when the optional fourth argument is nomatch
function addRanges() {
  local label="$1" ranges="$2" filter="$3" mode="${4:-}" cidr
  local verb="Adding" options=()

  if [[ "$mode" == "nomatch" ]]; then
    verb="Excluding"
    options+=(nomatch)
  fi

  while read -r cidr; do
    if [[ ! "$cidr" =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}/[0-9]{1,2}$ ]]; then
      echo "ERROR: Invalid CIDR range from $label IP ranges: $cidr" >&2
      exit 1
    fi
    echo "$verb $label range $cidr"
    ipset add allowed-domains "$cidr" "${options[@]}" -exist
  done < <(echo "$ranges" | jq -r "$filter" | aggregate -q)
}

# Fails when the first address of any IPv4 range the jq filter selects from the JSON is still allowed by the ipset
function checkExcluded() {
  local label="$1" ranges="$2" filter="$3" cidr

  while read -r cidr; do
    if ipset test allowed-domains "${cidr%/*}" >/dev/null 2>&1; then
      echo "ERROR: $label range $cidr is still allowed" >&2
      exit 1
    fi
  done < <(echo "$ranges" | jq -r "$filter" | aggregate -q)
}

# Fetches GitHub meta information and allows its IP ranges
function addGithubRanges() {
  local gh_ranges

  echo "Fetching GitHub IP ranges..."
  gh_ranges=$(fetchRanges "GitHub" https://api.github.com/meta '.web and .api and .git')

  echo "Processing GitHub IPs..."
  addRanges "GitHub" "$gh_ranges" '(.web + .api + .git)[]'
}

# Fetches Google's published IP ranges and allows them, except for the Google Cloud ranges
function addGoogleRanges() {
  local google_ranges cloud_ranges

  # Google serves proxy.golang.org and sum.golang.org from a large rotating address pool,
  # so single DNS lookups are not enough.
  # goog.json also contains Google Cloud ranges that anyone can rent, so the ranges published in
  # cloud.json are added afterwards as "nomatch" entries, which punch holes in the allowed blocks.
  # Google services that share front-end addresses (e.g. storage.googleapis.com) remain reachable;
  # an IP allowlist cannot tell them apart.
  echo "Fetching Google IP ranges..."
  google_ranges=$(fetchRanges "Google" https://www.gstatic.com/ipranges/goog.json '.prefixes')

  echo "Fetching Google Cloud IP ranges..."
  cloud_ranges=$(fetchRanges "Google Cloud" https://www.gstatic.com/ipranges/cloud.json '.prefixes')

  echo "Processing Google IPs..."
  addRanges "Google" "$google_ranges" '.prefixes[].ipv4Prefix // empty'

  # Must come after the allowed ranges: for a prefix present in both, the last add wins
  echo "Excluding Google Cloud IPs..."
  addRanges "Google Cloud" "$cloud_ranges" '.prefixes[].ipv4Prefix // empty' nomatch

  echo "Verifying Google Cloud IPs are excluded..."
  checkExcluded "Google Cloud" "$cloud_ranges" '.prefixes[].ipv4Prefix // empty'
}

# Resolves the other allowed domains and adds their IPs to the ipset
function addDomains() {
  local domain ips ip

  # Resolve and add other allowed domains
  for domain in \
    "registry.npmjs.org" \
    "api.anthropic.com" \
    "sentry.io" \
    "statsig.com" \
    "marketplace.visualstudio.com" \
    "vscode.blob.core.windows.net" \
    "update.code.visualstudio.com" \
    "proxy.golang.org" \
    "sum.golang.org" \
    "golang.org" \
    "go.dev" \
    "storage.googleapis.com"; do
    echo "Resolving $domain..."
    ips=$(dig +noall +answer A "$domain" | awk '$4 == "A" {print $5}')
    if [[ -z "$ips" ]]; then
      echo "ERROR: Failed to resolve $domain" >&2
      exit 1
    fi

    while read -r ip; do
      if [[ ! "$ip" =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
        echo "ERROR: Invalid IP from DNS for $domain: $ip" >&2
        exit 1
      fi
      echo "Adding $ip for $domain"
      ipset add allowed-domains "$ip" -exist
    done < <(echo "$ips")
  done
}

# Allows traffic to and from the network of the host
function allowHostNetwork() {
  local host_ip host_network

  # Get host IP from default route
  host_ip=$(ip route | grep default | cut -d" " -f3)
  if [[ -z "$host_ip" ]]; then
    echo "ERROR: Failed to detect host IP" >&2
    exit 1
  fi

  host_network=$(echo "$host_ip" | sed "s/\.[0-9]*$/.0\/24/")
  echo "Host network detected as: $host_network"

  # Set up remaining iptables rules
  iptables -A INPUT -s "$host_network" -j ACCEPT
  iptables -A OUTPUT -d "$host_network" -j ACCEPT
}

# Drops all traffic except the established connections and the allowed destinations
function lockDown() {
  # Set default policies to DROP first
  iptables -P INPUT DROP
  iptables -P FORWARD DROP
  iptables -P OUTPUT DROP

  # First allow established connections for already approved traffic
  iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
  iptables -A OUTPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

  # Then allow only specific outbound traffic to allowed domains
  iptables -A OUTPUT -m set --match-set allowed-domains dst -j ACCEPT

  # Explicitly REJECT all other outbound traffic for immediate feedback
  iptables -A OUTPUT -j REJECT --reject-with icmp-admin-prohibited

  echo "Firewall configuration complete"
}

# Fails unless an unlisted site is blocked and the GitHub API is reachable
function verify() {
  echo "Verifying firewall rules..."
  if curl --connect-timeout 5 https://example.com >/dev/null 2>&1; then
    echo "ERROR: Firewall verification failed - was able to reach https://example.com" >&2
    exit 1
  else
    echo "Firewall verification passed - unable to reach https://example.com as expected"
  fi

  # Verify GitHub API access
  if ! curl --connect-timeout 5 https://api.github.com/zen >/dev/null 2>&1; then
    echo "ERROR: Firewall verification failed - unable to reach https://api.github.com" >&2
    exit 1
  else
    echo "Firewall verification passed - able to reach https://api.github.com as expected"
  fi
}

# Locks the outbound traffic down to the allowlist and verifies it
function main() {
  # Fail closed: if the script exits before the policies are set in lockDown, drop all traffic
  trap 'iptables -P INPUT DROP; iptables -P FORWARD DROP; iptables -P OUTPUT DROP' EXIT

  flushRules
  allowBaseTraffic
  addGithubRanges
  addGoogleRanges
  addDomains
  allowHostNetwork
  lockDown
  verify
}

main
