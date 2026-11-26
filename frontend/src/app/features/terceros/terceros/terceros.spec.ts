import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { TercerosComponent } from './terceros';
import { environment } from '../../../../environments/environment';

describe('TercerosComponent', () => {
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    localStorage.clear();
    sessionStorage.clear();

    await TestBed.configureTestingModule({
      imports: [TercerosComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    }).compileComponents();

    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('debe cargar roles globales solo lectura y asignar uno o varios roles al crear un tercero', () => {
    const fixture = TestBed.createComponent(TercerosComponent);
    fixture.detectChanges();

    httpMock.expectOne(`${environment.apiUrl}/terceros/roles?activo=true`).flush({
      data: [
        { id: 1, codigo: 'CLIENTE', nombre_rol: 'Cliente', descripcion: null, activo: true, es_sistema: true },
        { id: 2, codigo: 'PROVEEDOR', nombre_rol: 'Proveedor', descripcion: null, activo: true, es_sistema: true },
        { id: 3, codigo: 'EMPLEADO', nombre_rol: 'Empleado', descripcion: null, activo: true, es_sistema: true },
      ],
    });
    httpMock.expectOne(`${environment.apiUrl}/terceros?page=1&page_size=15`).flush({
      data: [],
      page: 1,
      page_size: 15,
      total: 0,
    });

    const comp = fixture.componentInstance;
    expect(comp.loadingTerceros()).toBe(false);
    expect(comp.rolesDisponibles().length).toBe(3);

    // Crear tercero asignando múltiples roles (Cliente y Proveedor)
    comp.abrirNuevoTercero();
    comp.formTercero.patchValue({
      tipo_identificacion: 'NIT',
      numero_identificacion: '900123456-7',
      razon_social: 'Proveedora Industrial SAS',
      naturaleza_juridica: 'JURIDICA',
      email: 'ventas@proveedora.com',
      telefono: '3001112233',
      direccion: 'Calle 50 # 20-10',
      activo: true,
    });

    // Sin roles seleccionados no envía la petición
    comp.guardarTercero();
    httpMock.expectNone(`${environment.apiUrl}/terceros`);

    comp.toggleRolSeleccionado(1);
    comp.toggleRolSeleccionado(2);
    comp.guardarTercero();

    const postReq = httpMock.expectOne(`${environment.apiUrl}/terceros`);
    expect(postReq.request.method).toBe('POST');
    expect(postReq.request.body.rol_ids).toEqual([1, 2]);

    const creado = {
      id: 10,
      empresa_id: 1,
      tipo_identificacion: 'NIT' as const,
      numero_identificacion: '900123456-7',
      razon_social: 'Proveedora Industrial SAS',
      pais_id: null,
      naturaleza_juridica: 'JURIDICA' as const,
      email: 'ventas@proveedora.com',
      telefono: '3001112233',
      direccion: 'Calle 50 # 20-10',
      activo: true,
      roles: [
        { tercero_rol_id: 1, rol_id: 1, codigo: 'CLIENTE', nombre_rol: 'Cliente' },
        { tercero_rol_id: 2, rol_id: 2, codigo: 'PROVEEDOR', nombre_rol: 'Proveedor' },
      ],
      created_at: '2026-11-26T00:00:00Z',
      updated_at: '2026-11-26T00:00:00Z',
    };
    postReq.flush(creado);

    httpMock.expectOne(`${environment.apiUrl}/terceros?page=1&page_size=15`).flush({
      data: [creado],
      page: 1,
      page_size: 15,
      total: 1,
    });

    expect(comp.terceros().length).toBe(1);
    expect(comp.terceros()[0].roles.length).toBe(2);
  });
});
