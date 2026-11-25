import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideRouter } from '@angular/router';

import { MainLayoutComponent } from './main-layout';

describe('MainLayoutComponent', () => {
  beforeEach(async () => {
    localStorage.clear();
    sessionStorage.clear();

    await TestBed.configureTestingModule({
      imports: [MainLayoutComponent],
      providers: [provideHttpClient(), provideRouter([])],
    }).compileComponents();
  });

  it('debe desplegar el submenú de CRM con Resumen comercial y Leads al hacer clic', () => {
    const fixture = TestBed.createComponent(MainLayoutComponent);
    fixture.detectChanges();

    const el: HTMLElement = fixture.nativeElement;
    const crmButton = Array.from(el.querySelectorAll<HTMLButtonElement>('button.nav-toggle')).find(
      (btn) => btn.textContent?.includes('CRM'),
    );

    expect(crmButton).toBeTruthy();
    expect(el.querySelectorAll('.nav-subitem').length).toBe(0);

    crmButton!.click();
    fixture.detectChanges();

    const subitems = Array.from(el.querySelectorAll<HTMLElement>('.nav-subitem')).map((node) =>
      node.querySelector('.subitem-label')?.textContent?.trim(),
    );
    expect(subitems).toEqual([
      'Resumen comercial',
      'Leads',
      'Empresas y contactos',
      'Oportunidades',
      'Pipeline comercial',
      'Actividades',
      'Campañas',
      'Configuración',
    ]);
    expect(el.querySelectorAll('.nav-subitem.disabled').length).toBe(4);
    expect(el.querySelectorAll('a.nav-subitem').length).toBe(4);

    crmButton!.click();
    fixture.detectChanges();
    expect(el.querySelectorAll('.nav-subitem').length).toBe(0);
  });

  it('debe desplegar el submenú de Productos con Catálogo y Categorías activos y Configuración en Pronto', () => {
    const fixture = TestBed.createComponent(MainLayoutComponent);
    fixture.detectChanges();

    const el: HTMLElement = fixture.nativeElement;
    const prodButton = Array.from(el.querySelectorAll<HTMLButtonElement>('button.nav-toggle')).find(
      (btn) => btn.textContent?.includes('Productos'),
    );

    expect(prodButton).toBeTruthy();
    prodButton!.click();
    fixture.detectChanges();

    const subitems = Array.from(el.querySelectorAll<HTMLElement>('.nav-subitem')).map((node) =>
      node.querySelector('.subitem-label')?.textContent?.trim(),
    );
    expect(subitems).toEqual(['Catálogo', 'Categorías', 'Configuración']);
    expect(el.querySelectorAll('a.nav-subitem').length).toBe(2);
    expect(el.querySelectorAll('.nav-subitem.disabled').length).toBe(1);
  });
});
