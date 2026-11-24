import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideRouter } from '@angular/router';

import { DashboardComponent } from './dashboard';
import { environment } from '../../../environments/environment';

describe('DashboardComponent', () => {
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    localStorage.clear();
    sessionStorage.clear();

    await TestBed.configureTestingModule({
      imports: [DashboardComponent],
      providers: [provideHttpClient(), provideHttpClientTesting(), provideRouter([])],
    }).compileComponents();

    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('debe cargar los indicadores del Home desde el backend', () => {
    const fixture = TestBed.createComponent(DashboardComponent);
    fixture.detectChanges();

    httpMock.expectOne(`${environment.apiUrl}/crm/resumen?empresa_id=1`).flush({
      empresa_id: 1,
      total_leads: 12,
      leads_convertidos: 3,
      tasa_conversion_pct: 25,
      valor_estimado_total: 45000,
      actividades_pendientes: 4,
      actividades_vencidas: 1,
      por_etapa: [
        { estado_id: 1, nombre: 'Nuevo', orden: 1, es_estado_final: false, cantidad: 8, valor_estimado: 20000 },
        { estado_id: 2, nombre: 'Ganado', orden: 2, es_estado_final: true, cantidad: 4, valor_estimado: 25000 },
      ],
      por_temperatura: [{ temperatura: 'CALIENTE', cantidad: 5, valor_estimado: 30000 }],
    });

    httpMock
      .expectOne(`${environment.apiUrl}/crm/actividades?empresa_id=1&estado=PENDIENTE&page_size=5`)
      .flush({ data: [], total: 0 });

    httpMock
      .expectOne(`${environment.apiUrl}/usuarios?empresa_id=1&activo=true&page_size=1`)
      .flush({ data: [], total: 7 });

    httpMock
      .expectOne(`${environment.apiUrl}/empresas?activo=true&page_size=1`)
      .flush({ data: [], total: 2 });

    const component = fixture.componentInstance;
    expect(component.loading()).toBe(false);
    expect(component.data()?.backendDisponible).toBe(true);
    expect(component.data()?.crm?.total_leads).toBe(12);
    expect(component.data()?.totalUsuarios).toBe(7);
  });
});
