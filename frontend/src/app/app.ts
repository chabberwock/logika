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
    SelectLogFileDialog, SelectWorkspaceDirectoryDialog, Workspaces,
} from "../../wailsjs/go/app/App";
import {MatTabsModule} from "@angular/material/tabs";
import {MatDialog} from "@angular/material/dialog";
import {OnFileDrop} from "../../wailsjs/runtime";
import {MatButtonModule} from "@angular/material/button";
import {ErrorMessage} from "./error-message/error-message";


@Component({
    selector: 'app-root',
    imports: [RouterOutlet, Workspace, MatToolbarModule, MatIconModule, MatMenuModule, MatTabsModule, MatButtonModule],
    templateUrl: './app.html',
    styleUrl: './app.css'
})
export class App {
    protected readonly title = signal('logika-fe');
    public workspaces!: WorkspaceItem[];
    public dialog: MatDialog = inject(MatDialog);
    private cdr = inject(ChangeDetectorRef);

    async ngOnInit() {
        await this.updateWorkspaceList();
        OnFileDrop(this.handleFileDrop.bind(this), false);
    }

    async updateWorkspaceList() {
        let data = JSON.parse(await Workspaces());
        this.workspaces = [];
        if (data && Array.isArray(data)) {
            for (let i = 0; i < data.length; i++) {
                this.workspaces.push(data[i]);
            }
        }
        console.log("workspaces", this.workspaces);
    }

    async handleFileDrop(x: number, y: number, paths: string[]) {
        try {
            for (let i = 0; i < paths.length; i++) {
                if (!paths[i].endsWith(".log")) {
                    continue;
                }
                const workspacePath = await Import(paths[i]);
                let workspaceId = await Open(workspacePath);
                this.workspaces.push({id: workspaceId, dir: workspacePath});
                this.cdr.detectChanges();
            }

        } catch (e) {
            this.dialog.open(ErrorMessage, {data: {error: e, title: "Import error"}})
            // todo handle errors
        }
    }

    async importWorkspace() {
        try {
            const logFilePath = await SelectLogFileDialog();
            if (!logFilePath) {
                return;
            }
            const workspacePath = await Import(logFilePath);
            console.log('imported ' + workspacePath)
            let workspaceId = await Open(workspacePath);
            console.log('opened ' + workspaceId)
            this.workspaces.push({id: workspaceId, dir: workspacePath});
        } catch (e) {
            this.dialog.open(ErrorMessage, {data: {error: e, title: "Import error"}})
        }
    }

    async openWorkspace() {
        try {
            const workspacePath = await SelectWorkspaceDirectoryDialog();
            if (!workspacePath) {
                return;
            }
            let workspaceId = await Open(workspacePath);
            this.workspaces.push({id: workspaceId, dir: workspacePath});
        } catch (e) {
            this.dialog.open(ErrorMessage, {data: {error: e, title: "Open workspace error"}})
        }
    }

    async closeWorkspace(workspaceId: any) {
        for (let i = 0; i < this.workspaces.length; i++) {
            if (this.workspaces[i].id == workspaceId) {
                await Close(workspaceId);
                this.workspaces.splice(i, 1);
                return;
            }
        }
    }

    getLastPathSegment(p: string): string {
        if (!p) return '';
        const normalized = p.replace(/\\/g, '/');
        return normalized.split('/').filter(Boolean).pop() || '';
    }

    onTabChanged(idx: number) {
        setTimeout(() => {
            window.dispatchEvent(new Event('resize'));
        }, 100);
    }

}


export interface WorkspaceItem {
    id: string;
    dir: string;
}
