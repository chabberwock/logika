import {Directive, Input} from "@angular/core";

@Directive()
export abstract class FilterField {
    @Input() name!: string;
    @Input() value!: string;
    @Input() title!: string;
}