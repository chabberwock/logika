import {
    ChangeDetectorRef,
    Component, HostListener, inject,
    Input, NgZone,
    QueryList,
    ViewChild,
    ViewChildren,
    ViewEncapsulation
} from '@angular/core';
import {DynamicTableComponent} from "../../dynamic-table/dynamic-table";
import {
    CountLines,
    Filters, LoadSettings, LuaOutput, OpenWorkspaceDirectory,
    PreviewTemplate,
    Query, Reload, RunScript, SaveSettings,
} from "../../../wailsjs/go/app/App";
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
import {MatTooltipModule} from "@angular/material/tooltip";
import {MatDialog} from "@angular/material/dialog";
import {ErrorMessage} from "../error-message/error-message";
import {CdkMenuModule} from "@angular/cdk/menu";
import {ClipboardSetText} from "../../../wailsjs/runtime";
import {GetValue} from "../get-value/get-value";
import {first, firstValueFrom} from "rxjs";
import {AddBookmark, Bookmarks, DeleteBookmark} from "../../../wailsjs/go/app/Bookmarks";

@Component({
  selector: 'app-workspace',
  imports: [DynamicTableComponent, FilterContainer, MatSidenavModule,
      MatButtonModule, MatListModule, NgxJsonTreeviewComponent, ScrollingModule,
      MatButtonToggleModule, FormsModule, MatIconModule, NgxCodeJarComponent,
      MatTooltipModule,CdkMenuModule
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
    leftSideTab: string = "filters"
    readonly batchSize = 50;

    offset = 0;
    total = 0;
    loading = false;
    settings = "";
    output: string = "";
    console: string = "";
    consoleOutput: any = "";

    // for context menu
    selectedItem: any;

    bookmarks: any = {};

    public dialog: MatDialog = inject(MatDialog);
    private ngZone: NgZone = inject(NgZone);



    constructor(private cdr: ChangeDetectorRef) {}

    async ngAfterViewInit() {
        //this.viewport.checkViewportSize();
        await firstValueFrom(this.ngZone.onStable)
        window.dispatchEvent(new Event('resize'));
    }

    @HostListener('window:resize')
    onResize() {
        this.viewport.checkViewportSize();
    }

    highlightMethod(editor: CodeJarContainer) {
        if (editor.textContent !== null && editor.textContent !== undefined) {
            editor.innerHTML = hljs.highlight(editor.textContent, {
                language: 'lua'
            }).value;
        }
    }
    async ngOnInit() {
        this.listTemplate = await PreviewTemplate(this.id);
        await this.update();
        await this.refreshSettings();
        this.output = await LuaOutput(this.id);
        await this.updateBookmarks();
    }

    async update() {
        try {
            this.filters = JSON.parse(await Filters(this.id));
            const params = JSON.stringify(this.queryParams());

            this.data = JSON.parse(await Query(this.id, params, 0, 0));
            this.itemHeights = [];
            this.data.forEach((row: any) => {
                if (row == this.expandedElement) {
                    this.itemHeights.push(250)
                } else {
                    this.itemHeights.push(20);
                }
            });
            this.total = this.data.length;
            console.log("total: ", this.total);

        } catch (e) {
            this.dialog.open(ErrorMessage, {data: {error: e, title: "Could not update workspace"}})
        }
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

    async goToLine(line: any) {
        this.viewMode = "log";
        await firstValueFrom(this.ngZone.onStable)
        this.viewport.checkViewportSize();
        this.viewport.scrollToIndex(line - 1);
        this.expandedElement = this.data[line - 1];
    }

    collapsedView(row: any): string {
        if (this.listTemplate === undefined) {
            return 'loading...';
        }
        return this.listTemplate.replace(/\{\{([^{}]+)\}\}/g, (_, key) => {
            key = key.trim();
            if (key === 'text') {
                return JSON.stringify(row.data);
            }

            const path = key.split('.');
            let value = row;

            for (const part of path) {
                if (value && Object.prototype.hasOwnProperty.call(value, part)) {
                    value = value[part];
                } else {
                    value = '';
                    break;
                }
            }

            return value ?? '';
        });    }

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

    async openWorkspaceDirectory() {
        await OpenWorkspaceDirectory(this.id)
    }

    async runScript() {
        try {
            this.consoleOutput = await RunScript(this.id, this.console);
        } catch (e) {
            this.consoleOutput = e;
        }
        await this.updateBookmarks();
    }

    jsonParse(json: string) {
        return JSON.parse(json);
    }

    repairInfiniteScroll() {
        this.viewport.checkViewportSize();
    }

    async copyRow(selectedItem: any) {
        await ClipboardSetText(JSON.stringify(selectedItem.data, null, 2))
    }

    async addBookmark(selectedItem: any) {

        const dialogRef = this.dialog.open(GetValue, {
            data: { title: 'Add Bookmark', label: 'Name'}
        });

        let val = await firstValueFrom(dialogRef.afterClosed());
        await AddBookmark(this.id, String(selectedItem.line), val);
        await this.updateBookmarks();
    }

    async updateBookmarks() {
        this.bookmarks = await this.jsonParse(await Bookmarks(this.id));
    }

    async editBookmark(selectedItem: any) {
        const dialogRef = this.dialog.open(GetValue, {
            data: { title: 'Edit Bookmark', label: 'Name', value: this.bookmarks[selectedItem]}
        });
        let val = await firstValueFrom(dialogRef.afterClosed());
        await AddBookmark(this.id, String(selectedItem), val);
        await this.updateBookmarks();
    }

    async deleteBookmark(selectedItem: any) {
        await DeleteBookmark(this.id, selectedItem);
        await this.updateBookmarks();
    }

    openContextMenu(event: MouseEvent, item: any) {
        event.preventDefault();
        this.selectedItem = item;
    }


}
