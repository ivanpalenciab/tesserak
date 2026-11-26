import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { ComprasComponent } from './compras';
import { environment } from '../../../../environments/environment';

describe('ComprasComponent', () => {
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    localStorage.clear();
    sessionStorage.clear();

    await TestBed.configureTestingModule({
      imports: [ComprasComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    }).compileComponents();

    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('debe cargar terceros con rol PROVEEDOR, crear una compra seleccionando proveedor y eliminarla (soft delete)', () => {
    const fixture = TestBed.createComponent(ComprasComponent);
    fixture.detectChanges();

    httpMock.expectOne(`${environment.apiUrl}/crm/monedas`).flush({
      data: [{ id: 1, codigo_iso: 'COP', nombre_moneda: 'Peso colombiano', simbolo: '$' }],
    });
    httpMock.expectOne(`${environment.apiUrl}/productos?estado=true&page=1&page_size=100`).flush({
      data: [
        {
          id: 10,
          empresa_id: 1,
          codigo: 'SKU-01',
          nombre: 'Monitor 27"',
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
    httpMock.expectOne(`${environment.apiUrl}/terceros/roles?activo=true`).flush({
      data: [
        { id: 1, codigo: 'CLIENTE', nombre_rol: 'Cliente', descripcion: null, activo: true, es_sistema: true },
        { id: 2, codigo: 'PROVEEDOR', nombre_rol: 'Proveedor', descripcion: null, activo: true, es_sistema: true },
      ],
    });
    httpMock
      .expectOne(`${environment.apiUrl}/terceros?rol_id=2&activo=true&page=1&page_size=100`)
      .flush({
        data: [
          {
            id: 7,
            empresa_id: 1,
            tipo_identificacion: 'NIT',
            numero_identificacion: '900123456-7',
            razon_social: 'Tech Distro SAS',
            pais_id: null,
            naturaleza_juridica: 'JURIDICA',
            email: null,
            telefono: null,
            direccion: null,
            activo: true,
            roles: [
              { tercero_rol_id: 77, rol_id: 2, codigo: 'PROVEEDOR', nombre_rol: 'Proveedor' },
            ],
            created_at: '2026-01-01T00:00:00Z',
            updated_at: '2026-01-01T00:00:00Z',
          },
        ],
        page: 1,
        page_size: 100,
        total: 1,
      });

    httpMock.expectOne(`${environment.apiUrl}/compras?page=1&page_size=15`).flush({
      data: [],
      page: 1,
      page_size: 15,
      total: 0,
    });

    const comp = fixture.componentInstance;
    expect(comp.loading()).toBe(false);
    expect(comp.proveedores().length).toBe(1);
    expect(comp.proveedores()[0].tercero_rol_id).toBe(77);

    // Crear nueva compra seleccionando el proveedor del desplegable (sin ingresar consecutivo)
    comp.abrirNueva();
    comp.form.patchValue({
      numero_factura_proveedor: 'FAC-900',
      tercero_empresa_rol_id: 77,
      moneda_id: 1,
      estado: 'REGISTRADA',
      fecha_compra: '2026-11-26',
    });
    comp.detallesArray.at(0).patchValue({
      articulo_id: 10,
      descripcion: 'Monitor 27 pulgadas',
      cantidad: 2,
      costo_unitario: 500000,
      impuesto: 190000,
    });

    expect(comp.resumenFormulario()).toEqual({
      subtotal: 1000000,
      totalIva: 190000,
      total: 1190000,
    });

    comp.guardar();

    const postReq = httpMock.expectOne(`${environment.apiUrl}/compras`);
    expect(postReq.request.method).toBe('POST');
    expect(postReq.request.body.consecutivo).toBeUndefined();
    expect(postReq.request.body.tercero_empresa_rol_id).toBe(77);
    expect(postReq.request.body.proveedor_nombre).toBe('Tech Distro SAS');

    const compraCreada = {
      id: 50,
      empresa_id: 1,
      sucursal_id: 1,
      sucursal_nombre: 'Principal',
      tercero_empresa_rol_id: 77,
      proveedor_nombre: 'Tech Distro SAS',
      consecutivo: 'OC-2026-001',
      numero_factura_proveedor: 'FAC-900',
      moneda_id: 1,
      moneda_codigo: 'COP',
      moneda_simbolo: '$',
      estado: 'REGISTRADA' as const,
      fecha_compra: '2026-11-26T00:00:00Z',
      fecha_vencimiento: null,
      subtotal: 1000000,
      total_iva: 190000,
      total: 1190000,
      observaciones: null,
      usuario_id: 1,
      usuario_nombre: 'Admin',
      detalles: [],
      created_at: '2026-11-26T00:00:00Z',
      updated_at: '2026-11-26T00:00:00Z',
    };
    postReq.flush(compraCreada);

    httpMock.expectOne(`${environment.apiUrl}/compras?page=1&page_size=15`).flush({
      data: [compraCreada],
      page: 1,
      page_size: 15,
      total: 1,
    });

    expect(comp.compras().length).toBe(1);

    // Eliminar compra (soft delete)
    comp.solicitarEliminar(compraCreada);
    comp.confirmarEliminar();

    const delReq = httpMock.expectOne(`${environment.apiUrl}/compras/50`);
    expect(delReq.request.method).toBe('DELETE');
    delReq.flush(null);

    httpMock.expectOne(`${environment.apiUrl}/compras?page=1&page_size=15`).flush({
      data: [],
      page: 1,
      page_size: 15,
      total: 0,
    });

    expect(comp.compras().length).toBe(0);
  });
});
