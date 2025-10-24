import {ChangeDetectorRef, Component, inject, signal} from '@angular/core';
import {RouterOutlet} from '@angular/router';
import {Workspace} from "./workspace/workspace";
import {MatToolbarModule} from "@angular/material/toolbar";
import {MatIconModule} from "@angular/material/icon";
import {MatMenuModule} from "@angular/material/menu";
import {
    Close,
    Import,
    Open,
    Projects,
    SelectLogFileDialog,
    SelectProjectDirectoryDialog
} from "../../wailsjs/go/main/App";
import {MatTabsModule} from "@angular/material/tabs";
import {ImportProjectDialog} from "./import-project-dialog/import-project-dialog";
import {MatDialog} from "@angular/material/dialog";
import {firstValueFrom} from "rxjs";
import {OpenProjectDialog} from "./open-project-dialog/open-project-dialog";
import {OnFileDrop} from "../../wailsjs/runtime";
import {MatButtonModule} from "@angular/material/button";


@Component({
    selector: 'app-root',
    imports: [RouterOutlet, Workspace, MatToolbarModule, MatIconModule, MatMenuModule, MatTabsModule, MatButtonModule],
    templateUrl: './app.html',
    styleUrl: './app.css'
})
export class App {
    protected readonly title = signal('logika-fe');
    public projects!: Project[];
    public dialog: MatDialog = inject(MatDialog);
    private cdr = inject(ChangeDetectorRef);

    async ngOnInit() {
        await this.updateProjectList();
        OnFileDrop(this.handleFileDrop.bind(this), false);
    }

    async updateProjectList() {
        let data = JSON.parse(await Projects());
        this.projects = [];
        if (data && Array.isArray(data)) {
            for (let i = 0; i < data.length; i++) {
                this.projects.push(data[i]);
            }
        }
        console.log("projects", this.projects);
    }

    async handleFileDrop(x: number, y: number, paths: string[]) {
        try {
            for (let i = 0; i < paths.length; i++) {
                if (!paths[i].endsWith(".log")) {
                    continue;
                }
                const projectPath = await Import(paths[i]);
                let projectId = await Open(projectPath);
                this.projects.push({id: projectId, dir: projectPath});
                this.cdr.detectChanges();
            }

        } catch (e) {
            console.log(e);
            // todo handle errors
        }
    }

    async importProject() {
        try {
            const logFilePath = await SelectLogFileDialog();
            if (!logFilePath) {
                return;
            }
            const projectPath = await Import(logFilePath);
            let projectId = await Open(projectPath);
            this.projects.push({id: projectId, dir: projectPath});
        } catch (e) {
            console.log(e);
        }
    }

    async openProject() {
        try {
            const projectPath = await SelectProjectDirectoryDialog();
            if (!projectPath) {
                return;
            }
            let projectId = await Open(projectPath);
            this.projects.push({id: projectId, dir: projectPath});
        } catch (e) {
            console.log(e);
        }
    }

    async closeProject(projectId: any) {
        for (let i = 0; i < this.projects.length; i++) {
            if (this.projects[i].id == projectId) {
                await Close(projectId);
                this.projects.splice(i, 1);
                return;
            }
        }
    }

    getLastPathSegment(p: string): string {
        if (!p) return '';
        const normalized = p.replace(/\\/g, '/');
        return normalized.split('/').filter(Boolean).pop() || '';
    }

}


export interface Project {
    id: string;
    dir: string;
}
