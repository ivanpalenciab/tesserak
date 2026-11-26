import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { InventarioComponent } from './inventario';
import { environment } from '../../../../environments/environment';

describe('InventarioComponent', () => {
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    localStorage.clear();
    sessionStorage.clear();

    await TestBed.configureTestingModule({
      imports: [InventarioComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    }).compileComponents();

    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('debe mostrar primero solo ubicaciones, desplegar productos al hacer clic y permitir traslados entre ubicaciones', () => {
    const fixture = TestBed.createComponent(InventarioComponent);
    fixture.detectChanges();

    httpMock.expectOne(`${environment.apiUrl}/inventario/tipos-movimiento`).flush({
      data: [
        { id: 1, codigo: 'AJUSTE_POSITIVO', nombre: 'Ajuste positivo', naturaleza: 'ENTRADA', descripcion: null, activo: true },
        { id: 9, codigo: 'TRASLADO', nombre: 'Traslado entre ubicaciones', naturaleza: 'TRASLADO', descripcion: null, activo: true },
      ],
    });
    const bodega1 = {
      id: 1,
      empresa_id: 1,
      sucursal_id: 1,
      codigo: 'BOD-1',
      nombre: 'Bodega Principal',
      descripcion: null,
      activo: true,
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z',
    };
    const bodega2 = {
      id: 2,
      empresa_id: 1,
      sucursal_id: 1,
      codigo: 'VIT-1',
      nombre: 'Vitrina',
      descripcion: null,
      activo: true,
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z',
    };
    httpMock.expectOne(`${environment.apiUrl}/inventario/ubicaciones?activo=true`).flush({
      data: [bodega1, bodega2],
    });
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
    httpMock.expectOne(`${environment.apiUrl}/terceros?activo=true&page=1&page_size=100`).flush({
      data: [],
      page: 1,
      page_size: 100,
      total: 0,
    });

    const comp = fixture.componentInstance;
    expect(comp.loading()).toBe(false);
    expect(comp.ubicacionesActivas().length).toBe(2);
    // Al inicio NO se ha desplegado ninguna ubicación
    expect(comp.ubicacionSeleccionada()).toBeNull();

    // Al hacer clic en Bodega Principal carga sus productos desde inv_stock
    comp.seleccionarUbicacion(bodega1);

    httpMock
      .expectOne(`${environment.apiUrl}/inventario/stock?ubicacion_id=1&solo_con_stock=true&page=1&page_size=15`)
      .flush({
        data: [
          {
            id: 1,
            empresa_id: 1,
            articulo_id: 10,
            articulo_codigo: 'SKU-10',
            articulo_nombre: 'Teclado Mecánico',
            ubicacion_id: 1,
            ubicacion_codigo: 'BOD-1',
            ubicacion_nombre: 'Bodega Principal',
            cantidad: 25,
            costo_promedio: 120000,
            valor_total: 3000000,
            updated_at: '2026-11-26T00:00:00Z',
          },
        ],
        page: 1,
        page_size: 15,
        total: 1,
      });

    expect(comp.stockItems().length).toBe(1);
    expect(comp.stockItems()[0].cantidad).toBe(25);

    // Realizar un traslado interno de 5 unidades de Bodega Principal -> Vitrina
    comp.abrirNuevoMovimiento({ articulo_id: 10, ubicacion_origen_id: 1 });
    // Cambiar el tipo de movimiento al de traslado (simulando selección en el <select>)
    comp.formMov.controls.tipo_movimiento_id.setValue(9);
    comp.onCambiarTipoMovimiento();
    expect(comp.tipoMovSeleccionado()?.codigo).toBe('TRASLADO');

    comp.formMov.patchValue({ motivo: 'Traslado interno entre ubicaciones' });
    comp.detallesArray.at(0).patchValue({
      ubicacion_destino_id: 2,
      cantidad: 5,
    });
    comp.guardarMovimiento();

    const postMov = httpMock.expectOne(`${environment.apiUrl}/inventario/movimientos`);
    expect(postMov.request.method).toBe('POST');
    expect(postMov.request.body.tipo_movimiento_id).toBe(9);
    expect(postMov.request.body.detalles[0].ubicacion_origen_id).toBe(1);
    expect(postMov.request.body.detalles[0].ubicacion_destino_id).toBe(2);
    expect(postMov.request.body.detalles[0].cantidad).toBe(5);

    postMov.flush({
      id: 99,
      empresa_id: 1,
      sucursal_id: 1,
      tipo_movimiento_id: 9,
      tipo_movimiento_codigo: 'TRASLADO',
      tipo_movimiento_nombre: 'Traslado entre ubicaciones',
      tipo_movimiento_naturaleza: 'TRASLADO',
      consecutivo: 'INV-000001',
      fecha: '2026-11-26T00:00:00Z',
      tercero_id: null,
      motivo: 'Traslado interno entre ubicaciones',
      observaciones: null,
      usuario_id: 1,
      estado: 'APLICADO',
      detalles: [],
      created_at: '2026-11-26T00:00:00Z',
    });

    httpMock
      .expectOne(`${environment.apiUrl}/inventario/stock?ubicacion_id=1&solo_con_stock=true&page=1&page_size=15`)
      .flush({
        data: [
          {
            id: 1,
            empresa_id: 1,
            articulo_id: 10,
            articulo_codigo: 'SKU-10',
            articulo_nombre: 'Teclado Mecánico',
            ubicacion_id: 1,
            ubicacion_codigo: 'BOD-1',
            ubicacion_nombre: 'Bodega Principal',
            cantidad: 20,
            costo_promedio: 120000,
            valor_total: 2400000,
            updated_at: '2026-11-26T00:00:00Z',
          },
        ],
        page: 1,
        page_size: 15,
        total: 1,
      });

    expect(comp.stockItems()[0].cantidad).toBe(20);
  });
});
