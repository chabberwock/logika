import {Component, inject, signal} from '@angular/core';
import {
    MAT_DIALOG_DATA,
    MatDialogActions,
    MatDialogClose,
    MatDialogContent, MatDialogRef,
    MatDialogTitle
} from "@angular/material/dialog";
import {MatButtonModule} from "@angular/material/button";
import {MatIconModule} from "@angular/material/icon";
import {MatInputModule} from "@angular/material/input";
import {MatFormFieldModule} from "@angular/material/form-field";
import {FormsModule} from "@angular/forms";

@Component({
  selector: 'app-get-value',
    imports: [
        MatDialogTitle,
        MatDialogContent,
        MatDialogActions,
        MatButtonModule,
        MatDialogClose,
        MatIconModule,
        MatFormFieldModule,
        MatInputModule,
        FormsModule
    ],
  templateUrl: './get-value.html',
  styleUrl: './get-value.css'
})
export class GetValue {
    private dialogRef = inject(MatDialogRef<GetValue>);
    data = inject(MAT_DIALOG_DATA);
    value: string = '';

    constructor() {
        this.value = this.data.value? this.data.value : '';
    }

    cancel() {
        this.dialogRef.close();
    }

    submit() {
        this.dialogRef.close(this.value);
    }
}
