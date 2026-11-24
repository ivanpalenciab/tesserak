import { Component, inject } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

@Component({
  selector: 'app-crm-seccion-vacia',
  standalone: true,
  template: `
    <section class="crm-empty-page">
      <header class="page-header">
        <h1>{{ titulo }}</h1>
        @if (subtitulo) {
          <p class="subtitle">{{ subtitulo }}</p>
        }
      </header>
      <div class="empty-canvas"></div>
    </section>
  `,
  styles: `
    .crm-empty-page {
      display: flex;
      flex-direction: column;
      gap: 1.5rem;
    }

    .page-header {
      h1 {
        margin: 0 0 0.25rem;
        font-size: 1.65rem;
        font-weight: 700;
        letter-spacing: -0.025em;
        color: var(--text-primary);
      }
    }

    .subtitle {
      margin: 0;
      font-size: 0.9rem;
      color: var(--text-secondary);
    }

    .empty-canvas {
      min-height: 360px;
      border-radius: 14px;
      border: 1px dashed var(--border-color);
      background: var(--bg-surface);
    }
  `,
})
export class CrmSeccionVaciaComponent {
  private readonly route = inject(ActivatedRoute);

  readonly titulo = (this.route.snapshot.data['titulo'] as string) ?? 'CRM';
  readonly subtitulo = (this.route.snapshot.data['subtitulo'] as string) ?? '';
}
