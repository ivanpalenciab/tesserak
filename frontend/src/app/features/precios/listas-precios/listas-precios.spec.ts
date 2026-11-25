import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { ListasPreciosComponent } from './listas-precios';
import { environment } from '../../../../environments/environment';

describe('ListasPreciosComponent', () => {
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    localStorage.clear();
    sessionStorage.clear();

    await TestBed.configureTestingModule({
      imports: [ListasPreciosComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    }).compileComponents();

    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('debe mostrar primero las listas de precios y solo al hacer clic desplegar todos los artículos activos identificando los sin precio', () => {
    const fixture = TestBed.createComponent(ListasPreciosComponent);
    fixture.detectChanges();

    httpMock.expectOne(`${environment.apiUrl}/crm/monedas`).flush({
      data: [{ id: 1, codigo_iso: 'COP', nombre_moneda: 'Peso colombiano', simbolo: '$' }],
    });
    httpMock.expectOne(`${environment.apiUrl}/productos/unidades-medida`).flush({
      data: [{ id: 1, nombre: 'UND', descripcion: 'Unidad' }],
    });
    httpMock.expectOne(`${environment.apiUrl}/productos?estado=true&page=1&page_size=100`).flush({
      data: [
        {
          id: 10,
          empresa_id: 1,
          codigo: 'SKU-01',
          nombre: 'Portátil Pro',
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
        {
          id: 11,
          empresa_id: 1,
          codigo: 'SKU-02',
          nombre: 'Mouse Inalámbrico',
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
      total: 2,
    });

    const listaGeneral = {
      id: 1,
      empresa_id: 1,
      nombre: 'Lista General 2026',
      moneda_id: 1,
      moneda_codigo: 'COP',
      moneda_simbolo: '$',
      tipo_lista: 'GENERAL' as const,
      vigencia_desde: '2026-01-01',
      vigencia_hasta: null,
      estado: true,
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z',
    };

    httpMock.expectOne(`${environment.apiUrl}/listas-precios?page=1&page_size=10`).flush({
      data: [listaGeneral],
      page: 1,
      page_size: 10,
      total: 1,
    });

    const comp = fixture.componentInstance;
    expect(comp.loadingListas()).toBe(false);
    expect(comp.listas().length).toBe(1);
    // Al inicio NO hay lista seleccionada ni se despliega el detalle automáticamente
    expect(comp.listaSeleccionada()).toBeNull();

    // Al hacer clic en la lista se cargan los detalles y se cruzan con todos los artículos activos
    comp.seleccionarLista(listaGeneral);

    httpMock
      .expectOne(`${environment.apiUrl}/listas-precios/1/detalles?page=1&page_size=100`)
      .flush({
        data: [
          {
            id: 100,
            lista_precio_id: 1,
            articulo_id: 10,
            articulo_codigo: 'SKU-01',
            articulo_nombre: 'Portátil Pro',
            precio: 2500000,
            unidad_medida_id: 1,
            unidad_medida_nombre: 'UND',
            cantidad_minima: 1,
            vigencia_desde: null,
            vigencia_hasta: null,
            created_at: '2026-01-01T00:00:00Z',
            updated_at: '2026-01-01T00:00:00Z',
          },
        ],
        page: 1,
        page_size: 100,
        total: 1,
      });

    expect(comp.listaSeleccionada()?.id).toBe(1);
    expect(comp.filasArticulosPrecio().length).toBe(2);
    expect(comp.resumenAsignacion()).toEqual({ totalActivos: 2, conPrecio: 1, sinPrecio: 1 });

    // Filtrar solo los artículos sin precio asignado
    comp.onFiltroAsignacion('sin_precio');
    expect(comp.filasFiltradas().length).toBe(1);
    expect(comp.filasFiltradas()[0].articulo.id).toBe(11);
    expect(comp.filasFiltradas()[0].detalle).toBeNull();

    // Asignar precio al artículo faltante preseleccionado
    comp.abrirNuevoDetalle(comp.filasFiltradas()[0].articulo);
    expect(comp.formDetalle.controls.articulo_id.value).toBe(11);
    expect(comp.formDetalle.controls.unidad_medida_id.value).toBe(1);

    comp.formDetalle.patchValue({ precio: 85000, cantidad_minima: 1 });
    comp.guardarDetalle();

    const postDet = httpMock.expectOne(`${environment.apiUrl}/listas-precios/1/detalles`);
    expect(postDet.request.method).toBe('POST');
    expect(postDet.request.body.articulo_id).toBe(11);
    expect(postDet.request.body.precio).toBe(85000);
    postDet.flush({
      id: 101,
      lista_precio_id: 1,
      articulo_id: 11,
      articulo_codigo: 'SKU-02',
      articulo_nombre: 'Mouse Inalámbrico',
      precio: 85000,
      unidad_medida_id: 1,
      unidad_medida_nombre: 'UND',
      cantidad_minima: 1,
      vigencia_desde: null,
      vigencia_hasta: null,
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z',
    });

    httpMock
      .expectOne(`${environment.apiUrl}/listas-precios/1/detalles?page=1&page_size=100`)
      .flush({
        data: [
          {
            id: 100,
            lista_precio_id: 1,
            articulo_id: 10,
            articulo_codigo: 'SKU-01',
            articulo_nombre: 'Portátil Pro',
            precio: 2500000,
            unidad_medida_id: 1,
            unidad_medida_nombre: 'UND',
            cantidad_minima: 1,
            vigencia_desde: null,
            vigencia_hasta: null,
            created_at: '2026-01-01T00:00:00Z',
            updated_at: '2026-01-01T00:00:00Z',
          },
          {
            id: 101,
            lista_precio_id: 1,
            articulo_id: 11,
            articulo_codigo: 'SKU-02',
            articulo_nombre: 'Mouse Inalámbrico',
            precio: 85000,
            unidad_medida_id: 1,
            unidad_medida_nombre: 'UND',
            cantidad_minima: 1,
            vigencia_desde: null,
            vigencia_hasta: null,
            created_at: '2026-01-01T00:00:00Z',
            updated_at: '2026-01-01T00:00:00Z',
          },
        ],
        page: 1,
        page_size: 100,
        total: 2,
      });

    expect(comp.resumenAsignacion()).toEqual({ totalActivos: 2, conPrecio: 2, sinPrecio: 0 });
  });
});
