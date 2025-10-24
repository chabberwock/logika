import {
    ChangeDetectorRef,
    Component,
    Input,
    QueryList,
    ViewChild,
    ViewChildren,
    ViewEncapsulation
} from '@angular/core';
import {DynamicTableComponent} from "../../dynamic-table/dynamic-table";
import {
    CountLines,
    Filters, LoadSettings, LuaOutput, OpenProjectDirectory,
    PreviewTemplate,
    Query, Reload, RunScript, SaveSettings,
} from "../../../wailsjs/go/main/App";
import {MatSidenavModule} from "@angular/material/sidenav";
import {FilterContainer} from "../filter-container/filter-container";
import {OnFilter} from "../filter-container/events";
import {FilterField} from "../filter-container/filter-field";
import {MatButtonModule} from "@angular/material/button";
import {MatListModule} from "@angular/material/list";
import {NgxJsonTreeviewComponent} from "ngx-json-treeview";
import {CdkVirtualScrollViewport, ScrollingModule} from "@angular/cdk/scrolling";
import {MatButtonToggleGroup, MatButtonToggleModule} from "@angular/material/button-toggle";
import {FormsModule} from "@angular/forms";
import {MatIconModule} from "@angular/material/icon";
import hljs from 'highlight.js';
import {CodeJarContainer, NgxCodeJarComponent} from "ngx-codejar";

@Component({
  selector: 'app-workspace',
  imports: [DynamicTableComponent, FilterContainer, MatSidenavModule,
      MatButtonModule, MatListModule, NgxJsonTreeviewComponent, ScrollingModule,
      MatButtonToggleModule, FormsModule, MatIconModule, NgxCodeJarComponent
  ],
  templateUrl: './workspace.html',
  styleUrl: './workspace.css'
})
export class Workspace {
    @ViewChild(DynamicTableComponent) table!: DynamicTableComponent;
    @ViewChildren(FilterContainer) filterContainers!: QueryList<FilterContainer>;
    @ViewChild(CdkVirtualScrollViewport) viewport!: CdkVirtualScrollViewport;
    @Input() id!: string;
    data: any[] = [];
    rawData: any[] = [];
    itemHeights: number[] = [];
    filters: any;
    listTemplate!: string;
    expandedElement!: any;
    error!: any;
    viewMode: string = "log";
    readonly batchSize = 50;

    offset = 0;
    total = 0;
    loading = false;
    settings = "";
    output: string = "";
    console: string = "";
    consoleOutput: any = "";


    constructor(private cdr: ChangeDetectorRef) {}

    highlightMethod(editor: CodeJarContainer) {
        if (editor.textContent !== null && editor.textContent !== undefined) {
            editor.innerHTML = hljs.highlight(editor.textContent, {
                language: 'lua'
            }).value;
        }
    }
    async ngOnInit() {
        this.listTemplate = await PreviewTemplate(this.id);
        this.total = await CountLines(this.id, "{}");
        await this.update();
        await this.refreshSettings();
        this.output = await LuaOutput(this.id);
    }

    async update() {
        this.filters = JSON.parse(await Filters(this.id));
        const params = JSON.stringify(this.queryParams());

        this.rawData = JSON.parse(await Query(this.id, params, 0, 0));
        this.data = [];
        this.itemHeights = [];
        this.rawData.forEach((row: any) => {
            const {__text, __id, ...newRow} = row;
            this.data.push(newRow);
            if (row == this.expandedElement) {
                this.itemHeights.push(250)
            } else {
                this.itemHeights.push(20);
            }
        });
    }

    queryParams(): any {
        let params: any = {};
        this.filterContainers.forEach(c => {
            const val = c.value();
            if (val !== undefined) {
                params[c.name] = val;
            }
        });
        return params;
    }

    isExpanded(row: any): boolean {
        return row == this.expandedElement;
    }

    toggle(row: any) {
        this.expandedElement = this.isExpanded(row) ? null : row;
    }

    collapsedView(row: any): string {
        if (this.listTemplate === undefined) {
            return 'loading...';
        }
        return this.listTemplate.replace(/\{\{([^{}]+)\}\}/g, (_, key) => {
            return key in row ? row[key] : '';
        });
    }

    objectKeys(obj: object): string[] {
        if (typeof obj !== 'object') {
            return [];
        }
        return Object.keys(obj);
    }

    async refreshSettings() {
        this.settings = await LoadSettings(this.id);
    }

    async saveSettings() {
        await SaveSettings(this.id, this.settings);
        await Reload(this.id);
        this.output = await LuaOutput(this.id);
        this.listTemplate = await PreviewTemplate(this.id);
        this.total = await CountLines(this.id, "{}");
        await this.update();
    }

    async openProjectDirectory() {
        await OpenProjectDirectory(this.id)
    }

    async runScript() {
        try {
            this.consoleOutput = await RunScript(this.id, this.console);
        } catch (e) {
            this.consoleOutput = e;
        }
    }


}
