import { ComponentFixture, TestBed } from '@angular/core/testing';

import { FieldText } from './field-text';

describe('FieldText', () => {
  let component: FieldText;
  let fixture: ComponentFixture<FieldText>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [FieldText]
    })
    .compileComponents();

    fixture = TestBed.createComponent(FieldText);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
