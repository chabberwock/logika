import {Component, EventEmitter, Input, Output, QueryList, ViewChildren} from '@angular/core';
import {OnFilter} from "./events";
import {FilterField} from "./filter-field";
import {Filters} from "../../../wailsjs/go/app/App";
import {FieldText} from "./field-text/field-text";
import {FieldSelect} from "./field-select/field-select";

@Component({
  selector: 'app-filter-container',
  imports: [FieldText, FieldSelect],
  templateUrl: './filter-container.html',
  styleUrl: './filter-container.css'
})
export class FilterContainer {
    @Input() name!: string;
    @Input() settings!: any;
    @ViewChildren(FilterField) children!: QueryList<FilterField>;
    filters: any;

    value(): any {
        let resp: any = {};
        this.children.forEach(field => {
            if (field.value != undefined) {
                resp[field.name] = field.value;
            }
        });
        if (Object.keys(resp).length > 0) {
            return resp;
        }
    }

    objectKeys(obj: object): string[] {
        if (typeof obj !== 'object') {
            return [];
        }
        return Object.keys(obj);
    }

}
