{{- define "permissionRequire" -}}
&descriptor.PermissionRequire{Expression: {{ template "permissionExpression" .Expression }}}
{{- end }}

{{- define "permissionExpression" -}}
&descriptor.PermissionExpression{{ template "permissionExpressionValue" . }}
{{- end }}

{{- define "permissionExpressionValue" -}}
{ {{- template "permissionExpressionFields" . }} }
{{- end }}

{{- define "permissionExpressionFields" -}}
Mode: {{ permissionRequireLiteral .Mode }}{{ if .Code }}, Code: {{ quote .Code }}{{ end }}{{ if .Check }}, Check: &descriptor.PermissionCheckInvocation{ResourceSkelName: {{ quote .Check.ResourceSkelName }}, ActionName: {{ quote .Check.ActionName }}, CheckName: {{ quote .Check.CheckName }}, ServiceSkelName: {{ quote .Check.ServiceSkelName }}, MethodSkelName: {{ quote .Check.MethodSkelName }}, CodeArgumentName: {{ quote .Check.CodeArgumentName }}{{ if .Check.Arguments }}, Arguments: []*descriptor.PermissionCheckArgument{ {{- range $argument := .Check.Arguments }}{{ template "permissionCheckArgument" $argument }}, {{- end }} }{{ end }}}{{ end }}{{ if .Children }}, Children: []*descriptor.PermissionExpression{ {{- range $child := .Children }}{{ template "permissionExpressionValue" $child }}, {{- end }} }{{ end }}
{{- end }}

{{- define "permissionCheckArgument" -}}
{Name: {{ quote .Name }}, JsonPath: {{ quote .JsonPath }}, Type: {{ template "typeDescriptor" .Type }}}
{{- end }}
