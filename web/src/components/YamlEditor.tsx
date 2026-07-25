type Props = {
  value: string;
  onChange: (value: string) => void;
  rows?: number;
};

export function YamlEditor({ value, onChange, rows = 16 }: Props) {
  const lines = value.split('\n').length;

  return (
    <div className="yaml-editor">
      <div className="yaml-editor-toolbar">
        <span className="yaml-editor-badge">config.yaml</span>
        <span className="muted yaml-editor-meta">{lines} lines</span>
      </div>
      <textarea
        className="yaml-editor-input"
        rows={rows}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        spellCheck={false}
        aria-label="OTel config YAML"
      />
    </div>
  );
}
