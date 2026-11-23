import { Component, inject } from '@angular/core';

import { AuthService } from '../../core/auth/auth.service';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  template: `
    <div class="shell">
      <header class="topbar">
        <div class="brand">Tesserak</div>
        <div class="user-bar">
          @if (usuario(); as u) {
            <span class="user-name">{{ u.nombres }} {{ u.apellidos }}</span>
            <span class="user-meta">Empresa #{{ u.empresa_id }} · Sucursal #{{ u.sucursal_id }}</span>
          }
          <button type="button" class="btn-logout" (click)="logout()">Cerrar sesión</button>
        </div>
      </header>

      <main class="content">
        <h1>Bienvenido a Tesserak</h1>
        <p>Sesión iniciada correctamente. Aquí conectaremos los módulos de Empresas, Usuarios y CRM.</p>
      </main>
    </div>
  `,
  styles: `
    .shell {
      min-height: 100vh;
      background: var(--bg-page);
    }
    .topbar {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 1rem 2rem;
      background: var(--bg-surface);
      border-bottom: 1px solid var(--border-color);
    }
    .brand {
      font-weight: 700;
      font-size: 1.15rem;
      color: var(--text-primary);
    }
    .user-bar {
      display: flex;
      align-items: center;
      gap: 1rem;
    }
    .user-name {
      font-weight: 600;
      font-size: 0.92rem;
    }
    .user-meta {
      font-size: 0.82rem;
      color: var(--text-secondary);
    }
    .btn-logout {
      padding: 0.45rem 0.9rem;
      border: 1px solid var(--border-color);
      border-radius: 8px;
      background: #ffffff;
      color: var(--text-primary);
      font-size: 0.85rem;
      font-weight: 600;
      cursor: pointer;
    }
    .btn-logout:hover {
      background: #f1f5f9;
    }
    .content {
      max-width: 960px;
      margin: 3rem auto;
      padding: 0 1.5rem;
    }
    .content h1 {
      margin: 0 0 0.5rem;
      font-size: 1.65rem;
    }
    .content p {
      color: var(--text-secondary);
    }
  `,
})
export class DashboardComponent {
  private readonly auth = inject(AuthService);
  readonly usuario = this.auth.usuario;

  logout(): void {
    this.auth.logout();
  }
}
