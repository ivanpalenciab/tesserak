import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { CrmOportunidadesComponent } from './oportunidades';
import { environment } from '../../../../environments/environment';

describe('CrmOportunidadesComponent', () => {
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    localStorage.clear();
    sessionStorage.clear();

    await TestBed.configureTestingModule({
      imports: [CrmOportunidadesComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    }).compileComponents();

    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('debe cargar etapas, catálogos y listado de oportunidades', () => {
    const fixture = TestBed.createComponent(CrmOportunidadesComponent);
    fixture.detectChanges();

    httpMock
      .expectOne(`${environment.apiUrl}/crm/etapas-oportunidad?empresa_id=1&activo=true`)
      .flush({
        data: [
          {
            id: 5,
            empresa_id: 1,
            nombre: 'Propuesta',
            orden: 1,
            probabilidad_pct: 60,
            es_ganada: false,
            es_perdida: false,
            activo: true,
          },
        ],
      });
    httpMock
      .expectOne(`${environment.apiUrl}/crm/leads?empresa_id=1&page=1&page_size=100`)
      .flush({ data: [], page: 1, page_size: 100, total: 0 });
    httpMock
      .expectOne(`${environment.apiUrl}/crm/campanas?page_size=100`)
      .flush({ data: [], page: 1, page_size: 100, total: 0 });
    httpMock
      .expectOne(`${environment.apiUrl}/crm/motivos-descarte?empresa_id=1&activo=true`)
      .flush({ data: [] });

    httpMock
      .expectOne(`${environment.apiUrl}/crm/oportunidades?empresa_id=1&page=1&page_size=15`)
      .flush({
        data: [
          {
            id: 301,
            empresa_id: 1,
            sucursal_id: 1,
            lead_id: null,
            campana_id: null,
            estado_oportunidad_id: 5,
            motivo_descarte_id: null,
            usuario_asignado_id: null,
            titulo: 'Plan Enterprise Anual',
            descripcion: null,
            nombre_cliente: 'Acme SAS',
            valor_estimado: 18000,
            moneda_id: 1,
            probabilidad_pct: 60,
            fecha_cierre_estimada: '2026-12-15',
            fecha_cierre_real: null,
            prioridad: 'ALTA',
            created_by: 1,
            created_at: '2026-11-24T00:00:00Z',
            updated_at: '2026-11-24T00:00:00Z',
          },
        ],
        page: 1,
        page_size: 15,
        total: 1,
      });

    fixture.detectChanges();

    const comp = fixture.componentInstance;
    expect(comp.loading()).toBe(false);
    expect(comp.oportunidades().length).toBe(1);
    expect(comp.nombreEtapa(5)).toBe('Propuesta');

    const el: HTMLElement = fixture.nativeElement;
    const btnNueva = Array.from(el.querySelectorAll<HTMLButtonElement>('button')).find((b) =>
      b.textContent?.includes('Nueva oportunidad'),
    );
    expect(btnNueva).toBeTruthy();
    btnNueva!.click();
    fixture.detectChanges();
    expect(comp.modalAbierto()).toBe(true);
  });
});
