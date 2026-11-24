import { Component } from '@angular/core';

@Component({
  selector: 'app-resumen-comercial',
  standalone: true,
  template: `
    <section class="resumen-page">
      <header class="page-header">
        <h1>Resumen comercial</h1>
        <p class="subtitle">Espacio reservado para los indicadores y métricas comerciales del CRM.</p>
      </header>
      <div class="empty-canvas"></div>
    </section>
  `,
  styles: `
    .resumen-page {
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
export class ResumenComercialComponent {}
