import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideRouter, Router } from '@angular/router';

import { LoginComponent } from './login';
import { environment } from '../../../../environments/environment';

describe('LoginComponent', () => {
  let httpMock: HttpTestingController;
  let router: Router;

  beforeEach(async () => {
    localStorage.clear();
    sessionStorage.clear();

    await TestBed.configureTestingModule({
      imports: [LoginComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([{ path: 'dashboard', children: [] }]),
      ],
    }).compileComponents();

    httpMock = TestBed.inject(HttpTestingController);
    router = TestBed.inject(Router);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('debe rechazar el envío si el formulario es inválido', () => {
    const fixture = TestBed.createComponent(LoginComponent);
    const component = fixture.componentInstance;

    component.onSubmit();

    expect(component.form.invalid).toBe(true);
    expect(component.form.controls.identificador.touched).toBe(true);
    expect(component.form.controls.password.touched).toBe(true);
  });

  it('debe autenticar usando nombre de usuario y navegar al dashboard', () => {
    const fixture = TestBed.createComponent(LoginComponent);
    const component = fixture.componentInstance;
    const navigateSpy = vi.spyOn(router, 'navigate').mockResolvedValue(true);

    component.form.setValue({
      identificador: 'jperez',
      password: 'Secreta123',
      identificador_fiscal: '',
      recordar: true,
    });

    component.onSubmit();

    const req = httpMock.expectOne(`${environment.apiUrl}/auth/login`);
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({
      identificador: 'jperez',
      password: 'Secreta123',
    });

    req.flush({
      access_token: 'jwt-test-token',
      token_type: 'Bearer',
      expires_in: 43200,
      usuario: {
        id: 1,
        empresa_id: 1,
        sucursal_id: 1,
        nombre_usuario: 'jperez',
        nombres: 'Juan',
        apellidos: 'Pérez',
        email: 'juan@tesserak.com',
      },
    });

    expect(localStorage.getItem('tesserak_token')).toBe('jwt-test-token');
    expect(navigateSpy).toHaveBeenCalledWith(['/dashboard']);
  });

  it('debe autenticar también usando correo electrónico', () => {
    const fixture = TestBed.createComponent(LoginComponent);
    const component = fixture.componentInstance;
    const navigateSpy = vi.spyOn(router, 'navigate').mockResolvedValue(true);

    component.form.setValue({
      identificador: 'juan@tesserak.com',
      password: 'Secreta123',
      identificador_fiscal: '',
      recordar: false,
    });

    component.onSubmit();

    const req = httpMock.expectOne(`${environment.apiUrl}/auth/login`);
    expect(req.request.body).toEqual({
      identificador: 'juan@tesserak.com',
      password: 'Secreta123',
    });

    req.flush({
      access_token: 'jwt-email-token',
      token_type: 'Bearer',
      expires_in: 43200,
      usuario: {
        id: 1,
        empresa_id: 1,
        sucursal_id: 1,
        nombre_usuario: 'jperez',
        nombres: 'Juan',
        apellidos: 'Pérez',
        email: 'juan@tesserak.com',
      },
    });

    expect(sessionStorage.getItem('tesserak_token')).toBe('jwt-email-token');
    expect(navigateSpy).toHaveBeenCalledWith(['/dashboard']);
  });

  it('debe mostrar el campo de identificador fiscal si el backend responde empresa_requerida (409)', () => {
    const fixture = TestBed.createComponent(LoginComponent);
    const component = fixture.componentInstance;

    component.form.setValue({
      identificador: 'compartido@tesserak.com',
      password: 'Secreta123',
      identificador_fiscal: '',
      recordar: false,
    });

    component.onSubmit();

    const req = httpMock.expectOne(`${environment.apiUrl}/auth/login`);
    req.flush(
      { error: { code: 'empresa_requerida', message: 'indique empresa' } },
      { status: 409, statusText: 'Conflict' },
    );

    expect(component.showEmpresaField()).toBe(true);
    expect(component.errorMessage()).toContain('varias empresas');
  });
});
