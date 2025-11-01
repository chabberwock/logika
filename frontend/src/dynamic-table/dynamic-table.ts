import {Component, ElementRef, Input, OnInit, QueryList, ViewChildren} from '@angular/core';
import {Query} from "../../wailsjs/go/app/App";
import {MatTableModule} from "@angular/material/table";
import {Workspace} from "../app/workspace/workspace";

@Component({
    selector: 'app-dynamic-table',
    templateUrl: './dynamic-table.html',
    styleUrls: ['./dynamic-table.css'],
    imports: [
        MatTableModule
    ]
})
export class DynamicTableComponent implements OnInit {
    @ViewChildren('tableRow', { read: ElementRef }) rows!: QueryList<ElementRef<HTMLTableRowElement>>;
    displayedColumns: string[] = [];
    dataSource: any[] = [];
    loading = true;
    error: string | null = null;
    scrollToRow(index: number) {
        const row = this.rows.get(index);
        if (row) {
            row.nativeElement.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
    }

    update(data: any, displayedColumns: string[]) {
        if (Array.isArray(data) && data.length > 0) {
            this.dataSource = data;
            this.displayedColumns = displayedColumns;
            this.error = null;
        } else {
            this.error = 'Нет данных для отображения.';
        }
        this.loading = false;

    }

    async ngOnInit() {
    }

}
