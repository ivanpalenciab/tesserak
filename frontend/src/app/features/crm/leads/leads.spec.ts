import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { CrmLeadsComponent } from './leads';
import { environment } from '../../../../environments/environment';

describe('CrmLeadsComponent', () => {
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    localStorage.clear();
    sessionStorage.clear();

    await TestBed.configureTestingModule({
      imports: [CrmLeadsComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    }).compileComponents();

    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('debe cargar los catálogos y el listado de leads desde el backend', () => {
    const fixture = TestBed.createComponent(CrmLeadsComponent);
    fixture.detectChanges();

    httpMock
      .expectOne(`${environment.apiUrl}/crm/origenes?empresa_id=1&activo=true`)
      .flush({ data: [{ id: 1, empresa_id: 1, nombre: 'Web', activo: true }] });

    httpMock
      .expectOne(`${environment.apiUrl}/crm/estados?empresa_id=1`)
      .flush({
        data: [{ id: 10, empresa_id: 1, nombre: 'Nuevo', orden: 1, es_estado_final: false }],
      });

    httpMock
      .expectOne(`${environment.apiUrl}/crm/motivos-descarte?empresa_id=1&activo=true`)
      .flush({ data: [] });

    httpMock
      .expectOne(`${environment.apiUrl}/crm/campanas?page_size=100`)
      .flush({
        data: [
          {
            id: 5,
            empresa_id: 1,
            tipo_campana_id: 1,
            origen_id: 1,
            moneda_id: 1,
            nombre: 'Campaña Lanzamiento',
            estado: 'ACTIVA',
            fecha_inicio: '2026-11-01',
            fecha_fin: null,
          },
        ],
        page: 1,
        page_size: 100,
        total: 1,
      });

    httpMock
      .expectOne(`${environment.apiUrl}/crm/leads?empresa_id=1&page=1&page_size=15`)
      .flush({
        data: [
          {
            id: 101,
            empresa_id: 1,
            sucursal_id: 1,
            origen_id: 1,
            estado_id: 10,
            motivo_descarte_id: null,
            usuario_asignado_id: null,
            nombre: 'Laura',
            apellido: 'Gómez',
            nombre_empresa: 'Tech Corp',
            cargo: 'Directora',
            email: 'laura@techcorp.com',
            telefono: '3001234567',
            usuario_facebook: 'laura.gomez',
            temperatura: 'CALIENTE',
            score: 85,
            valor_estimado: 12000,
            moneda_id: null,
            convertido: false,
            convertido_tercero_empresa_rol_id: null,
            fecha_conversion: null,
            created_by: 1,
            created_at: '2026-11-20T10:00:00Z',
            updated_at: '2026-11-20T10:00:00Z',
          },
        ],
        page: 1,
        page_size: 15,
        total: 1,
      });

    fixture.detectChanges();

    const component = fixture.componentInstance;
    expect(component.loading()).toBe(false);
    expect(component.leads().length).toBe(1);
    expect(component.leads()[0].nombre).toBe('Laura');
    expect(component.nombreOrigen(1)).toBe('Web');
    expect(component.nombreEstado(10)).toBe('Nuevo');
    expect(component.campanasHabilitadas().length).toBe(1);
  });

  it('debe asociar el lead a la campaña seleccionada al crearlo', () => {
    const fixture = TestBed.createComponent(CrmLeadsComponent);
    fixture.detectChanges();

    httpMock
      .expectOne(`${environment.apiUrl}/crm/origenes?empresa_id=1&activo=true`)
      .flush({ data: [{ id: 1, empresa_id: 1, nombre: 'Web', activo: true }] });
    httpMock
      .expectOne(`${environment.apiUrl}/crm/estados?empresa_id=1`)
      .flush({
        data: [{ id: 10, empresa_id: 1, nombre: 'Nuevo', orden: 1, es_estado_final: false }],
      });
    httpMock
      .expectOne(`${environment.apiUrl}/crm/motivos-descarte?empresa_id=1&activo=true`)
      .flush({ data: [] });
    httpMock
      .expectOne(`${environment.apiUrl}/crm/campanas?page_size=100`)
      .flush({
        data: [
          {
            id: 5,
            empresa_id: 1,
            tipo_campana_id: 1,
            origen_id: 1,
            moneda_id: 1,
            nombre: 'Campaña Lanzamiento',
            estado: 'ACTIVA',
            fecha_inicio: '2026-11-01',
            fecha_fin: null,
          },
        ],
        page: 1,
        page_size: 100,
        total: 1,
      });
    httpMock
      .expectOne(`${environment.apiUrl}/crm/leads?empresa_id=1&page=1&page_size=15`)
      .flush({ data: [], page: 1, page_size: 15, total: 0 });

    const component = fixture.componentInstance;
    component.abrirNuevoLead();
    component.leadForm.patchValue({
      nombre: 'Carlos',
      apellido: 'Ruiz',
      origen_id: 1,
      estado_id: 10,
      campana_id: 5,
      tipo_atribucion: 'first_touch',
      utm_source: 'facebook',
      utm_campaign: 'lanzamiento',
    });

    component.guardarLead();

    const createReq = httpMock.expectOne(`${environment.apiUrl}/crm/leads`);
    expect(createReq.request.method).toBe('POST');
    createReq.flush({ id: 202, empresa_id: 1, origen_id: 1, estado_id: 10, nombre: 'Carlos', apellido: 'Ruiz' });

    const assocReq = httpMock.expectOne(`${environment.apiUrl}/crm/campanas/5/leads`);
    expect(assocReq.request.method).toBe('POST');
    expect(assocReq.request.body).toEqual({
      lead_id: 202,
      tipo_atribucion: 'first_touch',
      utm_source: 'facebook',
      utm_medium: undefined,
      utm_campaign: 'lanzamiento',
    });
    assocReq.flush({ id: 1, campana_id: 5, lead_id: 202 });

    httpMock
      .expectOne(`${environment.apiUrl}/crm/leads?empresa_id=1&page=1&page_size=15`)
      .flush({ data: [], page: 1, page_size: 15, total: 0 });

    expect(component.modalAbierto()).toBe(false);
    expect(component.exitoMensaje()).toContain('asociado a la campaña');
  });
});
