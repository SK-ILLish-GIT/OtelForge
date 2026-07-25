type Props = {
  title: string;
  hint?: string;
  action?: React.ReactNode;
};

export function EmptyState({ title, hint, action }: Props) {
  return (
    <div className="empty-state">
      <p className="empty-title">{title}</p>
      {hint && <p className="muted">{hint}</p>}
      {action}
    </div>
  );
}
