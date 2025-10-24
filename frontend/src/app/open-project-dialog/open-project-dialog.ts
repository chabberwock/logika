import {Component, inject} from '@angular/core';
import {
    MatDialogActions,
    MatDialogClose,
    MatDialogContent,
    MatDialogRef,
    MatDialogTitle
} from "@angular/material/dialog";
import {SelectLogFileDialog, SelectProjectDirectoryDialog} from "../../../wailsjs/go/main/App";

@Component({
  selector: 'app-open-project-dialog',
  imports: [
      MatDialogTitle,
      MatDialogContent,
      MatDialogActions,
      MatDialogClose,

  ],
  templateUrl: './open-project-dialog.html',
  styleUrl: './open-project-dialog.css'
})
export class OpenProjectDialog {
    logFilePath!: string;
    projectPath!: string;
    readonly dialogRef = inject(MatDialogRef<OpenProjectDialog>);
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
