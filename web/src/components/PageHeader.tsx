import type { ReactNode } from 'react';

type Props = {
  title: string;
  subtitle?: string;
  back?: ReactNode;
  actions?: ReactNode;
};

export function PageHeader({ title, subtitle, back, actions }: Props) {
  return (
    <header className="page-header">
      <div className="page-header-main">
        {back}
        <div>
          <h1>{title}</h1>
          {subtitle && <p className="page-subtitle">{subtitle}</p>}
        </div>
      </div>
      {actions && <div className="page-header-actions">{actions}</div>}
    </header>
  );
}
