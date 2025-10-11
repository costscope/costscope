package costscope.sbom

# Allow unless deny rule triggers
default allow = true

# Example: disallow GPL-3.0 licensed components (adjust per org policy)
deny[msg] {
  some c
  comp := input.components[c]
  some l
  comp.licenses[l].license.id == "GPL-3.0-only"
  msg := sprintf("GPL component blocked: %s %s", [comp.name, comp.version])
}

allow = false { count(deny) > 0 }
