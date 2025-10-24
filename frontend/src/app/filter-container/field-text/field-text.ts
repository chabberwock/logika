import {Component, forwardRef} from '@angular/core';
import {MatFormFieldModule} from "@angular/material/form-field";
import {MatInputModule} from "@angular/material/input";
import {FormsModule} from "@angular/forms";
import {MatButtonModule} from "@angular/material/button";
import {MatIconModule} from "@angular/material/icon";
import {FilterField} from "../filter-field";

@Component({
    selector: 'app-field-text',
    imports: [MatFormFieldModule, MatInputModule, FormsModule, MatButtonModule, MatIconModule],
    providers: [
        {provide: FilterField, useExisting: forwardRef(() => FieldText)}
    ],

    templateUrl: './field-text.html',
    styleUrl: './field-text.css'
})
export class FieldText extends FilterField {
}
