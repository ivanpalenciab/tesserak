import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { CrmCampanasComponent } from './campanas';
import { environment } from '../../../../environments/environment';

describe('CrmCampanasComponent', () => {
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    localStorage.clear();
    sessionStorage.clear();

    await TestBed.configureTestingModule({
      imports: [CrmCampanasComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    }).compileComponents();

    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('debe cargar tipos de campaña, orígenes y listado de campañas', () => {
    const fixture = TestBed.createComponent(CrmCampanasComponent);
    fixture.detectChanges();

    httpMock.expectOne(`${environment.apiUrl}/crm/tipos-campana`).flush({
      data: [
        {
          id: 1,
          empresa_id: 1,
          nombre: 'Publicidad digital',
          descripcion: 'Ads',
          activo: true,
          created_at: '2026-11-01T00:00:00Z',
          updated_at: '2026-11-01T00:00:00Z',
        },
      ],
    });

    httpMock
      .expectOne(`${environment.apiUrl}/crm/origenes?empresa_id=1&activo=true`)
      .flush({ data: [{ id: 2, empresa_id: 1, nombre: 'Facebook', activo: true }] });

    httpMock.expectOne(`${environment.apiUrl}/crm/monedas`).flush({
      data: [{ id: 1, codigo_iso: 'USD', nombre_moneda: 'Dólar estadounidense', simbolo: '$' }],
    });

    httpMock
      .expectOne(`${environment.apiUrl}/crm/leads?empresa_id=1&page=1&page_size=100`)
      .flush({ data: [], page: 1, page_size: 100, total: 0 });

    httpMock.expectOne(`${environment.apiUrl}/crm/campanas?page=1&page_size=15`).flush({
      data: [
        {
          id: 10,
          empresa_id: 1,
          tipo_campana_id: 1,
          origen_id: 2,
          moneda_id: 1,
          nombre: 'Black Friday 2026',
          descripcion: 'Campaña Q4',
          objetivo: '200 leads',
          fecha_inicio: '2026-11-01',
          fecha_fin: '2026-11-30',
          presupuesto: 5000,
          estado: 'ACTIVA',
          created_by: 1,
          created_at: '2026-11-01T00:00:00Z',
          updated_at: '2026-11-01T00:00:00Z',
        },
      ],
      page: 1,
      page_size: 15,
      total: 1,
    });

    fixture.detectChanges();

    const comp = fixture.componentInstance;
    expect(comp.loading()).toBe(false);
    expect(comp.campanas().length).toBe(1);
    expect(comp.nombreTipo(1)).toBe('Publicidad digital');
    expect(comp.nombreOrigen(2)).toBe('Facebook');
  });

  it('debe cargar indicadores, leads y gastos al abrir el detalle de una campaña', () => {
    const fixture = TestBed.createComponent(CrmCampanasComponent);
    fixture.detectChanges();

    httpMock.expectOne(`${environment.apiUrl}/crm/tipos-campana`).flush({ data: [] });
    httpMock
      .expectOne(`${environment.apiUrl}/crm/origenes?empresa_id=1&activo=true`)
      .flush({ data: [] });
    httpMock.expectOne(`${environment.apiUrl}/crm/monedas`).flush({ data: [] });
    httpMock
      .expectOne(`${environment.apiUrl}/crm/leads?empresa_id=1&page=1&page_size=100`)
      .flush({ data: [], page: 1, page_size: 100, total: 0 });
    httpMock
      .expectOne(`${environment.apiUrl}/crm/campanas?page=1&page_size=15`)
      .flush({ data: [], page: 1, page_size: 15, total: 0 });

    const comp = fixture.componentInstance;
    comp.abrirDetalle({
      id: 10,
      empresa_id: 1,
      tipo_campana_id: 1,
      origen_id: null,
      moneda_id: 1,
      nombre: 'Promo Q4',
      descripcion: null,
      objetivo: null,
      fecha_inicio: '2026-11-01',
      fecha_fin: null,
      presupuesto: 3000,
      estado: 'ACTIVA',
      created_by: 1,
      created_at: '2026-11-01T00:00:00Z',
      updated_at: '2026-11-01T00:00:00Z',
    });

    httpMock.expectOne(`${environment.apiUrl}/crm/campanas/10/indicadores`).flush({
      campana_id: 10,
      empresa_id: 1,
      nombre: 'Promo Q4',
      estado: 'ACTIVA',
      moneda_id: 1,
      presupuesto: 3000,
      total_leads: 4,
      gasto_acumulado: 1200,
      presupuesto_disponible: 1800,
      gastos_por_moneda: [{ moneda_id: 1, total_gastado: 1200, cantidad_gastos: 2 }],
    });
    httpMock
      .expectOne(`${environment.apiUrl}/crm/campanas/10/leads?page=1&page_size=50`)
      .flush({ data: [], page: 1, page_size: 50, total: 0 });
    httpMock
      .expectOne(`${environment.apiUrl}/crm/campanas/10/gastos?page=1&page_size=50`)
      .flush({ data: [], page: 1, page_size: 50, total: 0 });

    expect(comp.indicadores()?.total_leads).toBe(4);
    expect(comp.indicadores()?.presupuesto_disponible).toBe(1800);
  });
});
