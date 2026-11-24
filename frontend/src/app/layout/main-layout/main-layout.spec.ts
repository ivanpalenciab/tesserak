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

    const subitems = Array.from(el.querySelectorAll<HTMLAnchorElement>('.nav-subitem')).map((a) =>
      a.textContent?.trim(),
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

    crmButton!.click();
    fixture.detectChanges();
    expect(el.querySelectorAll('.nav-subitem').length).toBe(0);
  });
});
