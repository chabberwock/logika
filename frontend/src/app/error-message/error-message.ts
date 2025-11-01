import {Component, inject} from '@angular/core';
import {
    MAT_DIALOG_DATA,
    MatDialogActions,
    MatDialogClose,
    MatDialogContent,
    MatDialogTitle
} from "@angular/material/dialog";
import {MatButtonModule} from "@angular/material/button";

@Component({
  selector: 'app-error-message',
  imports: [MatDialogTitle, MatDialogContent, MatDialogActions, MatButtonModule, MatDialogClose],
  templateUrl: './error-message.html',
  styleUrl: './error-message.css'
})
export class ErrorMessage {
    data = inject(MAT_DIALOG_DATA);
}
