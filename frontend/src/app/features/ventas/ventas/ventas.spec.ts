import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { VentasComponent, CONSUMIDOR_FINAL_DOC } from './ventas';
import { environment } from '../../../../environments/environment';

describe('VentasComponent', () => {
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    localStorage.clear();
    sessionStorage.clear();

    await TestBed.configureTestingModule({
      imports: [VentasComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    }).compileComponents();

    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('debe preseleccionar Consumidor Final (222222222222), registrar una venta y permitir soft delete', () => {
    const fixture = TestBed.createComponent(VentasComponent);
    fixture.detectChanges();

    httpMock.expectOne(`${environment.apiUrl}/productos?estado=true&page=1&page_size=100`).flush({
      data: [
        {
          id: 10,
          empresa_id: 1,
          codigo: 'SKU-10',
          nombre: 'Teclado Mecánico',
          descripcion: null,
          tipo_articulo_id: 1,
          unidad_medida_id: 1,
          categoria_id: 1,
          marca_id: null,
          grupo_impuesto_id: 1,
          estado: true,
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z',
        },
      ],
      page: 1,
      page_size: 100,
      total: 1,
    });
    httpMock.expectOne(`${environment.apiUrl}/productos/grupos-impuesto`).flush({
      data: [{ id: 1, empresa_id: 1, nombre: 'IVA 19%', tarifa: 19, activo: true }],
    });
    httpMock.expectOne(`${environment.apiUrl}/terceros?activo=true&page=1&page_size=100`).flush({
      data: [
        {
          id: 100,
          empresa_id: 1,
          tipo_identificacion: 'CC',
          numero_identificacion: CONSUMIDOR_FINAL_DOC,
          razon_social: 'Consumidor Final',
          pais_id: null,
          naturaleza_juridica: 'NATURAL',
          email: null,
          telefono: null,
          direccion: null,
          activo: true,
          roles: [{ id: 1, codigo: 'CLIENTE', nombre_rol: 'Cliente' }],
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z',
        },
        {
          id: 101,
          empresa_id: 1,
          tipo_identificacion: 'NIT',
          numero_identificacion: '900111222',
          razon_social: 'Proveedor ABC',
          pais_id: null,
          naturaleza_juridica: 'JURIDICA',
          email: null,
          telefono: null,
          direccion: null,
          activo: true,
          roles: [{ id: 2, codigo: 'PROVEEDOR', nombre_rol: 'Proveedor' }],
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z',
        },
      ],
      page: 1,
      page_size: 100,
      total: 2,
    });
    httpMock.expectOne(`${environment.apiUrl}/inventario/ubicaciones?activo=true`).flush({
      data: [
        {
          id: 1,
          empresa_id: 1,
          sucursal_id: 1,
          codigo: 'BOD-1',
          nombre: 'Bodega Principal',
          descripcion: null,
          activo: true,
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z',
        },
      ],
    });
    httpMock.expectOne(`${environment.apiUrl}/ventas?page=1&page_size=15`).flush({
      data: [],
      page: 1,
      page_size: 15,
      total: 0,
    });

    const comp = fixture.componentInstance;
    expect(comp.clientes().length).toBe(1);
    expect(comp.clientes()[0].numero_identificacion).toBe(CONSUMIDOR_FINAL_DOC);

    comp.abrirNuevaVenta();
    expect(comp.form.controls.cliente_id.value).toBe(100);
    expect(comp.form.controls.ubicacion_id.value).toBe(1);

    comp.detallesArray.at(0).patchValue({
      articulo_id: 10,
      cantidad: 2,
      precio_unitario: 150000,
    });
    expect(comp.tarifaLinea(comp.detallesArray.at(0))).toBe(19);
    expect(comp.calcularIVA()).toBe(57000);
    expect(comp.calcularTotal()).toBe(357000);
    comp.guardarVenta();

    const postReq = httpMock.expectOne(`${environment.apiUrl}/ventas`);
    expect(postReq.request.method).toBe('POST');
    expect(postReq.request.body.cliente_id).toBe(100);
    expect(postReq.request.body.ubicacion_id).toBe(1);
    expect(postReq.request.body.detalles[0].cantidad).toBe(2);
    expect(postReq.request.body.detalles[0].impuesto).toBe(57000);

    const creada = {
      id: 50,
      empresa_id: 1,
      sucursal_id: 1,
      ubicacion_id: 1,
      ubicacion_nombre: 'Bodega Principal',
      cliente_id: 100,
      cliente_nombre: 'Consumidor Final',
      cliente_identificacion: 'CC 222222222222',
      consecutivo: 'VEN-000001',
      moneda_id: null,
      estado: 'REGISTRADA' as const,
      fecha_venta: '2026-11-26T00:00:00Z',
      fecha_vencimiento: null,
      subtotal: 300000,
      total_iva: 57000,
      total: 357000,
      observaciones: null,
      usuario_id: 1,
      created_at: '2026-11-26T00:00:00Z',
      updated_at: '2026-11-26T00:00:00Z',
    };
    postReq.flush(creada);

    httpMock.expectOne(`${environment.apiUrl}/ventas?page=1&page_size=15`).flush({
      data: [creada],
      page: 1,
      page_size: 15,
      total: 1,
    });
    expect(comp.ventas().length).toBe(1);

    // Soft delete
    comp.solicitarEliminar(creada);
    comp.confirmarEliminar();
    const delReq = httpMock.expectOne(`${environment.apiUrl}/ventas/50`);
    expect(delReq.request.method).toBe('DELETE');
    delReq.flush(null);

    httpMock.expectOne(`${environment.apiUrl}/ventas?page=1&page_size=15`).flush({
      data: [],
      page: 1,
      page_size: 15,
      total: 0,
    });
    expect(comp.ventas().length).toBe(0);
  });
});
