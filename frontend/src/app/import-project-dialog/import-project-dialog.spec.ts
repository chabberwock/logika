import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ImportProjectDialog } from './import-project-dialog';

describe('ImportProjectDialog', () => {
  let component: ImportProjectDialog;
  let fixture: ComponentFixture<ImportProjectDialog>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ImportProjectDialog]
    })
    .compileComponents();

    fixture = TestBed.createComponent(ImportProjectDialog);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
