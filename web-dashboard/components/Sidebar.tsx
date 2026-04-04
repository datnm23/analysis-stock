'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

export default function Sidebar() {
  const [isOpen, setIsOpen] = useState(false);
  const pathname = usePathname();

  // Close sidebar on route change
  useEffect(() => {
    setIsOpen(false);
  }, [pathname]);

  return (
    <>
      {/* Mobile Header */}
      <div className="mobile-header">
        <div className="sidebar-logo">
          <span>📊</span>
          <h1>VNStock</h1>
        </div>
        <button className="menu-toggle" onClick={() => setIsOpen(!isOpen)}>
          {isOpen ? '✕' : '☰'}
        </button>
      </div>

      {/* Overlay */}
      {isOpen && <div className="sidebar-overlay" onClick={() => setIsOpen(false)} />}

      {/* Sidebar */}
      <aside className={`sidebar ${isOpen ? 'open' : ''}`}>
        <div className="sidebar-logo hidden-mobile">
          <span>📊</span>
          <h1>VNStock</h1>
        </div>
        <nav>
          <ul className="nav-links">
            <li>
              <Link href="/" className={pathname === '/' ? 'active' : ''}>
                🏠 Tổng Quan
              </Link>
            </li>
            <li>
              <Link href="/analysis" className={pathname === '/analysis' ? 'active' : ''}>
                🔍 Phân Tích
              </Link>
            </li>
            <li>
              <Link href="/stock/FPT" className={pathname?.startsWith('/stock') ? 'active' : ''}>
                📈 Chi Tiết
              </Link>
            </li>
            <li>
              <Link href="/reports" className={pathname === '/reports' ? 'active' : ''}>
                📋 Báo Cáo
              </Link>
            </li>
          </ul>
        </nav>
      </aside>
    </>
  );
}
