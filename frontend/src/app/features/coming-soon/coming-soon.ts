import { Component, inject } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';

@Component({
  selector: 'app-coming-soon',
  standalone: true,
  imports: [RouterLink],
  template: `
    <section class="placeholder">
      <span class="badge">Módulo conectado al backend</span>
      <h1>{{ titulo }}</h1>
      <p>{{ descripcion }}</p>
      <a class="btn-back" routerLink="/dashboard">← Volver al inicio</a>
    </section>
  `,
  styles: `
    .placeholder {
      background: var(--bg-surface);
      border: 1px solid var(--border-color);
      border-radius: 16px;
      padding: 2.5rem;
      max-width: 620px;
    }
    .badge {
      display: inline-block;
      padding: 0.25rem 0.65rem;
      border-radius: 999px;
      background: #eef2ff;
      color: #3730a3;
      font-size: 0.75rem;
      font-weight: 600;
      margin-bottom: 0.9rem;
    }
    h1 {
      margin: 0 0 0.5rem;
      font-size: 1.5rem;
    }
    p {
      margin: 0 0 1.5rem;
      color: var(--text-secondary);
      line-height: 1.6;
    }
    .btn-back {
      display: inline-block;
      padding: 0.55rem 1rem;
      border-radius: 9px;
      background: var(--primary);
      color: #ffffff;
      text-decoration: none;
      font-size: 0.86rem;
      font-weight: 600;
    }
  `,
})
export class ComingSoonComponent {
  private readonly route = inject(ActivatedRoute);
  readonly titulo = (this.route.snapshot.data['titulo'] as string) ?? 'Módulo';
  readonly descripcion =
    (this.route.snapshot.data['descripcion'] as string) ??
    'La API de este módulo ya está lista; en el siguiente paso construiremos esta pantalla.';
}
