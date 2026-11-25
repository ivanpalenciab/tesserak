import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { ProductosCatalogoComponent } from './catalogo';
import { environment } from '../../../../environments/environment';

describe('ProductosCatalogoComponent', () => {
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    localStorage.clear();
    sessionStorage.clear();

    await TestBed.configureTestingModule({
      imports: [ProductosCatalogoComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    }).compileComponents();

    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('debe cargar catálogos y crear un servicio con duración en horas', () => {
    const fixture = TestBed.createComponent(ProductosCatalogoComponent);
    fixture.detectChanges();

    httpMock.expectOne(`${environment.apiUrl}/productos/tipos-articulo`).flush({
      data: [
        { id: 1, codigo: 'PRODUCTO', nombre: 'Producto' },
        { id: 2, codigo: 'SERVICIO', nombre: 'Servicio' },
      ],
    });
    httpMock.expectOne(`${environment.apiUrl}/productos/metodos-valoracion`).flush({
      data: [{ id: 1, codigo: 'FIFO', nombre: 'FIFO' }],
    });
    httpMock.expectOne(`${environment.apiUrl}/productos/modalidades-servicio`).flush({
      data: [{ id: 1, codigo: 'POR_HORA', nombre: 'Por hora' }],
    });
    httpMock.expectOne(`${environment.apiUrl}/productos/grupos-impuesto?activo=true`).flush({
      data: [{ id: 1, empresa_id: 1, nombre: 'IVA 19%', tarifa: 19, activo: true }],
    });
    httpMock
      .expectOne(`${environment.apiUrl}/productos/unidades-medida`)
      .flush({ data: [{ id: 1, nombre: 'UND', descripcion: 'Unidad' }] });
    httpMock
      .expectOne(`${environment.apiUrl}/productos/categorias?activo=true`)
      .flush({ data: [{ id: 10, empresa_id: 1, nombre: 'Consultoría', activo: true }] });
    httpMock
      .expectOne(`${environment.apiUrl}/productos/marcas?activo=true`)
      .flush({ data: [] });

    httpMock.expectOne(`${environment.apiUrl}/productos?page=1&page_size=15`).flush({
      data: [],
      page: 1,
      page_size: 15,
      total: 0,
    });

    const comp = fixture.componentInstance;
    expect(comp.loading()).toBe(false);

    comp.abrirNuevo();
    expect(comp.tipoSeleccionadoCodigo()).toBeNull();

    comp.seleccionarTipo(2);
    expect(comp.tipoSeleccionadoCodigo()).toBe('SERVICIO');

    comp.form.patchValue({
      codigo: 'SRV-001',
      nombre: 'Soporte técnico',
      unidad_medida_id: 1,
      categoria_id: 10,
      grupo_impuesto_id: 1,
      modalidad_id: 1,
      duracion_estimada: 2.5,
      requiere_programacion: true,
      control_consumos: false,
      estado: true,
    });

    comp.guardar();

    const postReq = httpMock.expectOne(`${environment.apiUrl}/productos`);
    expect(postReq.request.method).toBe('POST');
    expect(postReq.request.body.grupo_impuesto_id).toBe(1);
    expect(postReq.request.body.servicio).toEqual({
      modalidad_id: 1,
      duracion_estimada: 2.5,
      requiere_programacion: true,
      control_consumos: false,
    });
    postReq.flush({ id: 5, empresa_id: 1, codigo: 'SRV-001', nombre: 'Soporte técnico', estado: true });

    httpMock.expectOne(`${environment.apiUrl}/productos?page=1&page_size=15`).flush({
      data: [],
      page: 1,
      page_size: 15,
      total: 0,
    });
  });
});
