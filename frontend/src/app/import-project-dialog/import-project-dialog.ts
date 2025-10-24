import {Component, inject} from '@angular/core';
import {
    MAT_DIALOG_DATA,
    MatDialogActions, MatDialogClose,
    MatDialogContent,
    MatDialogRef,
    MatDialogTitle
} from "@angular/material/dialog";
import {SelectLogFileDialog, SelectProjectDirectoryDialog} from "../../../wailsjs/go/main/App";

@Component({
  selector: 'app-import-project-dialog',
  imports: [
      MatDialogTitle,
      MatDialogContent,
      MatDialogActions,
      MatDialogClose,
  ],
  templateUrl: './import-project-dialog.html',
  styleUrl: './import-project-dialog.css'
})
export class ImportProjectDialog {
    logFilePath!: string;
    projectPath!: string;
    readonly dialogRef = inject(MatDialogRef<ImportProjectDialog>);
    error!: any;

    async selectLogFile() {
        try {
            this.error = "";
            this.logFilePath = await SelectLogFileDialog();

        } catch (e) {
            this.error = e;
        }
    }

    async selectProject() {
        try {
            this.error = "";
            this.projectPath = await SelectProjectDirectoryDialog();

        } catch (e) {
            this.error = e;
        }
    }


    onCancel() {
        this.dialogRef.close();
    }

    onSave() {
        this.dialogRef.close({logFilePath: this.logFilePath, projectPath: this.projectPath});
    }




}
