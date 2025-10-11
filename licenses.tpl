{{- /*
  Template: licenses.tpl
  Purpose: Used by go-licenses (optional) to emit a simple CSV style summary of licenses.
  Fields exposed (each element):
    Name          – module path
    Version       – semantic version (may be pseudo)
    LicenseNames  – slice of detected license identifiers
  This template is intentionally minimal; authoritative license texts remain in upstream modules.
*/ -}}
Module,Version,Licenses
{{- range . }}
{{ .Name }},{{ .Version }},{{ join .LicenseNames "|" }}
{{- end }}
## License Report for CostScope

| Package | License |
|---------|---------|
{{- range . }}
| {{.Name}} | {{.LicenseName}} |
{{- end }}
