{{- define "permissionRequire" -}}
&skel.PermRequire{Expr: {{ template "permissionExpr" .Expr }}}
{{- end }}

{{- define "permissionExpr" -}}
&skel.PermExpr{{ template "permissionExprValue" . }}
{{- end }}

{{- define "permissionExprValue" -}}
{ {{- template "permissionExprFields" . }} }
{{- end }}

{{- define "permissionExprFields" -}}
Mode: {{ permissionRequireLiteral .Mode }}{{ if .Code }}, Code: {{ quote .Code }}{{ end }}{{ if .Check }}, Check: &skel.PermCheckInvocation{ResourceSkelName: {{ quote .Check.ResourceSkelName }}, ActionName: {{ quote .Check.ActionName }}, CheckName: {{ quote .Check.CheckName }}, ServiceSkelName: {{ quote .Check.ServiceSkelName }}, MethodSkelName: {{ quote .Check.MethodSkelName }}, CodeArgumentName: {{ quote .Check.CodeArgumentName }}{{ if .Check.Arguments }}, Arguments: []*skel.PermCheckArgument{ {{- range $argument := .Check.Arguments }}{{ template "permCheckArgument" $argument }}, {{- end }} }{{ end }}}{{ end }}{{ if .Children }}, Children: []*skel.PermExpr{ {{- range $child := .Children }}{{ template "permissionExprValue" $child }}, {{- end }} }{{ end }}
{{- end }}

{{- define "permCheckArgument" -}}
{Name: {{ quote .Name }}, JsonPath: {{ quote .JsonPath }}, Type: {{ template "typeSchema" .Type }}}
{{- end }}
