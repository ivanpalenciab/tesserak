import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { ProductosConfiguracionComponent } from './configuracion';
import { environment } from '../../../../environments/environment';

describe('ProductosConfiguracionComponent', () => {
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    localStorage.clear();
    sessionStorage.clear();

    await TestBed.configureTestingModule({
      imports: [ProductosConfiguracionComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    }).compileComponents();

    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('debe cargar marcas e impuestos y permitir crear un grupo de impuesto', () => {
    const fixture = TestBed.createComponent(ProductosConfiguracionComponent);
    fixture.detectChanges();

    httpMock.expectOne(`${environment.apiUrl}/productos/marcas`).flush({
      data: [{ id: 1, empresa_id: 1, nombre: 'Tesserak', activo: true, created_at: '', updated_at: '' }],
    });
    httpMock.expectOne(`${environment.apiUrl}/productos/grupos-impuesto`).flush({
      data: [{ id: 1, empresa_id: 1, nombre: 'Exento (0%)', tarifa: 0, activo: true }],
    });

    const comp = fixture.componentInstance;
    expect(comp.loading()).toBe(false);
    expect(comp.marcas().length).toBe(1);
    expect(comp.impuestos().length).toBe(1);

    comp.tabActiva.set('impuestos');
    comp.abrirNuevoImpuesto();
    comp.formImpuesto.patchValue({ nombre: 'IVA 19%', tarifa: 19, activo: true });
    comp.guardarImpuesto();

    const postReq = httpMock.expectOne(`${environment.apiUrl}/productos/grupos-impuesto`);
    expect(postReq.request.method).toBe('POST');
    expect(postReq.request.body).toEqual({ nombre: 'IVA 19%', tarifa: 19, activo: true });
    postReq.flush({ id: 2, empresa_id: 1, nombre: 'IVA 19%', tarifa: 19, activo: true });

    httpMock.expectOne(`${environment.apiUrl}/productos/grupos-impuesto`).flush({
      data: [
        { id: 1, empresa_id: 1, nombre: 'Exento (0%)', tarifa: 0, activo: true },
        { id: 2, empresa_id: 1, nombre: 'IVA 19%', tarifa: 19, activo: true },
      ],
    });
    expect(comp.impuestos().length).toBe(2);
  });
});
