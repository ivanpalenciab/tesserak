import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { ProductosCategoriasComponent } from './categorias';
import { environment } from '../../../../environments/environment';

describe('ProductosCategoriasComponent', () => {
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    localStorage.clear();
    sessionStorage.clear();

    await TestBed.configureTestingModule({
      imports: [ProductosCategoriasComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    }).compileComponents();

    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('debe listar, crear, activar/desactivar y eliminar (soft delete) categorías', () => {
    const fixture = TestBed.createComponent(ProductosCategoriasComponent);
    fixture.detectChanges();

    httpMock.expectOne(`${environment.apiUrl}/productos/categorias`).flush({
      data: [{ id: 1, empresa_id: 1, nombre: 'Tecnología', activo: true }],
    });

    const comp = fixture.componentInstance;
    expect(comp.categorias().length).toBe(1);

    // Desactivar
    comp.toggleActivo(comp.categorias()[0]);
    const patchReq = httpMock.expectOne(`${environment.apiUrl}/productos/categorias/1`);
    expect(patchReq.request.method).toBe('PATCH');
    patchReq.flush({ id: 1, empresa_id: 1, nombre: 'Tecnología', activo: false });
    expect(comp.categorias()[0].activo).toBe(false);

    // Crear
    comp.abrirNueva();
    comp.form.patchValue({ nombre: 'Servicios', activo: true });
    comp.guardar();
    const postReq = httpMock.expectOne(`${environment.apiUrl}/productos/categorias`);
    expect(postReq.request.method).toBe('POST');
    postReq.flush({ id: 2, empresa_id: 1, nombre: 'Servicios', activo: true });
    httpMock.expectOne(`${environment.apiUrl}/productos/categorias`).flush({ data: [] });

    // Soft delete
    comp.solicitarEliminar({
      id: 2,
      empresa_id: 1,
      nombre: 'Servicios',
      activo: true,
      created_at: '',
      updated_at: '',
    });
    comp.confirmarEliminar();
    const delReq = httpMock.expectOne(`${environment.apiUrl}/productos/categorias/2`);
    expect(delReq.request.method).toBe('DELETE');
    delReq.flush(null);
    httpMock.expectOne(`${environment.apiUrl}/productos/categorias`).flush({ data: [] });
  });
});
