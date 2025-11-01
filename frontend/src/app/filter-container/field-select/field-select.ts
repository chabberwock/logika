import {Component, forwardRef, Input} from '@angular/core';
import {FilterField} from "../filter-field";
import {MatFormFieldModule} from "@angular/material/form-field";
import {MatSelectModule} from "@angular/material/select";
import {FormsModule} from "@angular/forms";

@Component({
  selector: 'app-field-select',
  imports: [MatFormFieldModule, MatSelectModule, FormsModule],
    providers: [
        {provide: FilterField, useExisting: forwardRef(() => FieldSelect)}
    ],

  templateUrl: './field-select.html',
  styleUrl: './field-select.css'
})

export class FieldSelect extends FilterField{
    @Input() options!: Option[];
}


interface Option {
    title: string;
    value: string;
}
