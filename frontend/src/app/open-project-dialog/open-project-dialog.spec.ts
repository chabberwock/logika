import { ComponentFixture, TestBed } from '@angular/core/testing';

import { OpenProjectDialog } from './open-project-dialog';

describe('OpenProjectDialog', () => {
  let component: OpenProjectDialog;
  let fixture: ComponentFixture<OpenProjectDialog>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [OpenProjectDialog]
    })
    .compileComponents();

    fixture = TestBed.createComponent(OpenProjectDialog);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
