import { ComponentFixture, TestBed } from '@angular/core/testing';

import { FilterContainer } from './filter-container';

describe('FilterContainer', () => {
  let component: FilterContainer;
  let fixture: ComponentFixture<FilterContainer>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [FilterContainer]
    })
    .compileComponents();

    fixture = TestBed.createComponent(FilterContainer);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
