import { ComponentFixture, TestBed } from '@angular/core/testing';

import { GetValue } from './get-value';

describe('GetValue', () => {
  let component: GetValue;
  let fixture: ComponentFixture<GetValue>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [GetValue]
    })
    .compileComponents();

    fixture = TestBed.createComponent(GetValue);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
