import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { signal } from '@angular/core';

import { GestionEmpresaComponent } from './gestion-empresa';
import { AuthService } from '../../../core/auth/auth.service';
import { environment } from '../../../../environments/environment';

describe('GestionEmpresaComponent', () => {
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    localStorage.clear();
    sessionStorage.clear();

    const authStub = {
      usuario: signal({
        id: 5,
        empresa_id: 1,
        sucursal_id: 1,
        nombre_usuario: 'admin',
        nombres: 'Admin',
        apellidos: 'Tesserak',
        email: 'admin@tesserak.local',
      }),
    };

    await TestBed.configureTestingModule({
      imports: [GestionEmpresaComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        { provide: AuthService, useValue: authStub },
      ],
    }).compileComponents();

    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('debe permitir modificar prefijos de compras/ventas y ver/cambiar la key del usuario', () => {
    const fixture = TestBed.createComponent(GestionEmpresaComponent);
    fixture.detectChanges();

    httpMock.expectOne(`${environment.apiUrl}/empresas/1`).flush({
      id: 1,
      nombre_legal: 'Tesserak SAS',
      razon_social: 'Tesserak SAS',
      identificador_fiscal: '900123456-7',
      email: 'info@tesserak.com',
      telefono: null,
      direccion: null,
      activo: true,
    });
    httpMock.expectOne(`${environment.apiUrl}/usuarios/5`).flush({
      id: 5,
      key: '11111111-2222-4333-8444-555555555555',
      empresa_id: 1,
      sucursal_id: 1,
      nombre_usuario: 'admin',
      nombres: 'Admin',
      apellidos: 'Tesserak',
      email: 'admin@tesserak.local',
      activo: true,
    });
    httpMock.expectOne(`${environment.apiUrl}/consecutivos/tipos-documento`).flush({
      data: [
        { id: 1, codigo: 'COMPRA', nombre: 'Compras', descripcion: null, prefijo_defecto: 'COM', activo: true },
        { id: 2, codigo: 'VENTA', nombre: 'Ventas', descripcion: null, prefijo_defecto: 'VEN', activo: true },
      ],
    });
    httpMock.expectOne(`${environment.apiUrl}/consecutivos/secuencias`).flush({
      data: [
        {
          id: 10,
          empresa_id: 1,
          tipo_documento_id: 1,
          tipo_documento_codigo: 'COMPRA',
          tipo_documento_nombre: 'Compras',
          prefijo: 'COM',
          siguiente_numero: 5,
          longitud_relleno: 6,
          frecuencia_reinicio: 'NUNCA',
          clave_periodo: null,
          activo: true,
          ejemplo_proximo: 'COM-000005',
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z',
        },
      ],
    });

    const comp = fixture.componentInstance;
    expect(comp.loading()).toBe(false);
    expect(comp.secuenciasItems().length).toBe(2);
    expect(comp.usuarioDetalle()?.key).toBe('11111111-2222-4333-8444-555555555555');

    // 1. Modificar prefijo de Compras (PATCH porque ya existe secuenciaId = 10)
    const itemCompra = comp.secuenciasItems()[0];
    itemCompra.form.patchValue({ prefijo: 'OC' });
    comp.guardarSecuencia(itemCompra);

    const patchReq = httpMock.expectOne(`${environment.apiUrl}/consecutivos/secuencias/10`);
    expect(patchReq.request.method).toBe('PATCH');
    expect(patchReq.request.body.prefijo).toBe('OC');
    patchReq.flush({
      id: 10,
      empresa_id: 1,
      tipo_documento_id: 1,
      prefijo: 'OC',
      siguiente_numero: 5,
      longitud_relleno: 6,
      frecuencia_reinicio: 'NUNCA',
      clave_periodo: null,
      activo: true,
      ejemplo_proximo: 'OC-000005',
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z',
    });

    // 2. Guardar prefijo de Ventas (POST porque aún no tenía secuencia creada)
    const itemVenta = comp.secuenciasItems()[1];
    itemVenta.form.patchValue({ prefijo: 'FAC' });
    comp.guardarSecuencia(itemVenta);

    const postSec = httpMock.expectOne(`${environment.apiUrl}/consecutivos/secuencias`);
    expect(postSec.request.method).toBe('POST');
    expect(postSec.request.body.tipo_documento_id).toBe(2);
    expect(postSec.request.body.prefijo).toBe('FAC');
    postSec.flush({
      id: 11,
      empresa_id: 1,
      tipo_documento_id: 2,
      prefijo: 'FAC',
      siguiente_numero: 1,
      longitud_relleno: 6,
      frecuencia_reinicio: 'NUNCA',
      clave_periodo: null,
      activo: true,
      ejemplo_proximo: 'FAC-000001',
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z',
    });

    // 3. Cambiar (regenerar) la key del usuario
    comp.abrirConfirmarRegenerar();
    comp.confirmarRegenerarKey();

    const keyReq = httpMock.expectOne(`${environment.apiUrl}/usuarios/5/regenerar-key`);
    expect(keyReq.request.method).toBe('POST');
    keyReq.flush({
      id: 5,
      key: 'aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee',
      empresa_id: 1,
      sucursal_id: 1,
      nombre_usuario: 'admin',
      nombres: 'Admin',
      apellidos: 'Tesserak',
      email: 'admin@tesserak.local',
      activo: true,
    });

    expect(comp.usuarioDetalle()?.key).toBe('aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee');
    expect(comp.keyVisible()).toBe(true);
  });
});
