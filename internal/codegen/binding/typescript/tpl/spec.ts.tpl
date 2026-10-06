{{- if $.HasWire }}
import type { VrpcWireSchema } from '@yorun-ai/vrpc';
{{ end -}}
{{- range $factory := $.WireFactories }}
{{ $factory.Code }}
{{ end -}}
{{- range $service := $.Services }}
export const {{ $service.SpecName }} = {
  serviceName: '{{ $service.SkelName }}',
  methods: {
  {{- range $method := $service.Methods }}
    {{ $method.Name }}: '{{ $method.SkelName }}',
  {{- end }}
  },
  {{- if $service.WireMethods }}
  wire: {
    {{- range $method := $service.WireMethods }}
    {{ $method.Name }}: {
      {{- if $method.ArgumentsSchema }}
      arguments: {{ $method.ArgumentsSchema }} satisfies VrpcWireSchema,
      {{- end }}
      {{- if $method.ResultSchema }}
      result: {{ $method.ResultSchema }} satisfies VrpcWireSchema,
      {{- end }}
    },
    {{- end }}
  },
  {{- end }}
} as const;
{{ end -}}
{{- if and (not $.Services) (not $.HasWire) }}
export {};
{{- end }}
