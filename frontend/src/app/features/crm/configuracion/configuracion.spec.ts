import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { CrmConfiguracionComponent } from './configuracion';
import { environment } from '../../../../environments/environment';

describe('CrmConfiguracionComponent', () => {
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    localStorage.clear();
    sessionStorage.clear();

    await TestBed.configureTestingModule({
      imports: [CrmConfiguracionComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    }).compileComponents();

    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('debe cargar los 4 catálogos de configuración y permitir crear/eliminar con soft delete', () => {
    const fixture = TestBed.createComponent(CrmConfiguracionComponent);
    fixture.detectChanges();

    httpMock.expectOne(`${environment.apiUrl}/crm/origenes?empresa_id=1`).flush({
      data: [{ id: 1, empresa_id: 1, nombre: 'Web', activo: true }],
    });
    httpMock.expectOne(`${environment.apiUrl}/crm/estados?empresa_id=1`).flush({
      data: [{ id: 10, empresa_id: 1, nombre: 'Nuevo', orden: 1, es_estado_final: false }],
    });
    httpMock.expectOne(`${environment.apiUrl}/crm/etapas-oportunidad?empresa_id=1`).flush({
      data: [
        {
          id: 20,
          empresa_id: 1,
          nombre: 'Propuesta',
          orden: 1,
          probabilidad_pct: 50,
          es_ganada: false,
          es_perdida: false,
          activo: true,
        },
      ],
    });
    httpMock.expectOne(`${environment.apiUrl}/crm/motivos-descarte?empresa_id=1`).flush({
      data: [{ id: 30, empresa_id: 1, descripcion: 'Precio alto', activo: true }],
    });

    const comp = fixture.componentInstance;
    expect(comp.loading()).toBe(false);
    expect(comp.origenes().length).toBe(1);
    expect(comp.etapas().length).toBe(1);
    expect(comp.etapasOportunidad().length).toBe(1);
    expect(comp.motivos().length).toBe(1);

    // Crear etapa de oportunidad
    comp.abrirNuevaEtapaOp();
    comp.etapaOpForm.patchValue({
      nombre: 'Negociación',
      orden: 2,
      probabilidad_pct: 80,
      es_ganada: false,
      es_perdida: false,
      activo: true,
    });
    comp.guardarEtapaOp();

    const postEtapaOp = httpMock.expectOne(`${environment.apiUrl}/crm/etapas-oportunidad`);
    expect(postEtapaOp.request.method).toBe('POST');
    postEtapaOp.flush({
      id: 21,
      empresa_id: 1,
      nombre: 'Negociación',
      orden: 2,
      probabilidad_pct: 80,
      es_ganada: false,
      es_perdida: false,
      activo: true,
    });

    httpMock.expectOne(`${environment.apiUrl}/crm/etapas-oportunidad?empresa_id=1`).flush({
      data: [
        { id: 20, empresa_id: 1, nombre: 'Propuesta', orden: 1, probabilidad_pct: 50, es_ganada: false, es_perdida: false, activo: true },
        { id: 21, empresa_id: 1, nombre: 'Negociación', orden: 2, probabilidad_pct: 80, es_ganada: false, es_perdida: false, activo: true },
      ],
    });

    expect(comp.etapasOportunidad().length).toBe(2);

    // Soft delete de origen
    comp.solicitarEliminar('origenes', 1, 'Web');
    comp.confirmarEliminar();

    const delReq = httpMock.expectOne(`${environment.apiUrl}/crm/origenes/1`);
    expect(delReq.request.method).toBe('DELETE');
    delReq.flush(null);

    httpMock.expectOne(`${environment.apiUrl}/crm/origenes?empresa_id=1`).flush({ data: [] });
    expect(comp.origenes().length).toBe(0);
  });
});
