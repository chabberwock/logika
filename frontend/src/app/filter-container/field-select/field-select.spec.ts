import { ComponentFixture, TestBed } from '@angular/core/testing';

import { FieldSelect } from './field-select';

describe('FieldSelect', () => {
  let component: FieldSelect;
  let fixture: ComponentFixture<FieldSelect>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [FieldSelect]
    })
    .compileComponents();

    fixture = TestBed.createComponent(FieldSelect);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
