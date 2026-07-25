import { Link, NavLink, Outlet } from 'react-router-dom';
import { useAuth } from '../auth/AuthContext';

const links = [
  { to: '/instances', label: 'Instances', icon: '▦' },
  { to: '/events', label: 'Events', icon: '◎' },
  { to: '/launch', label: 'Launch', icon: '▶' },
];

export function Layout() {
  const { user, logout, isAdmin } = useAuth();

  return (
    <div className="app">
      <aside className="sidebar">
        <Link to="/instances" className="sidebar-brand">
          <span className="brand-mark">OF</span>
          OtelForge
        </Link>
        <nav className="sidebar-nav">
          {links.map((link) => (
            <NavLink key={link.to} to={link.to}>
              <span className="nav-icon" aria-hidden>{link.icon}</span>
              {link.label}
            </NavLink>
          ))}
          {isAdmin && (
            <NavLink to="/admin">
              <span className="nav-icon" aria-hidden>⚙</span>
              Admin
            </NavLink>
          )}
        </nav>
        <div className="sidebar-footer">
          <span className="sidebar-user">{user?.email}</span>
          <button type="button" className="btn-ghost" onClick={logout}>Sign out</button>
        </div>
      </aside>
      <main className="main">
        <Outlet />
      </main>
    </div>
  );
}
