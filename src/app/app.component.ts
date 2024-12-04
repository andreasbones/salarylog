import { Component, OnInit } from '@angular/core';
import { SalaryService, SalaryEntry } from './services/salary.service';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: []
})
export class AppComponent implements OnInit {
  newName: string = '';
  selectedName: string = '';
  salary: number | null = null;
  year: number | null = null;

  names: string[] = [];
  entries: SalaryEntry[] = [];

  constructor(private salaryService: SalaryService) {}

  ngOnInit() {
    this.salaryService.names$.subscribe(names => this.names = names);
    this.salaryService.entries$.subscribe(entries => this.entries = entries);
  }

  onNameSubmit() {
    if (this.newName.trim()) {
      this.salaryService.addName(this.newName.trim())
        .subscribe(names => {
          this.names = names;
          this.newName = '';
        });
    }
  }

  onEntrySubmit() {
    if (this.selectedName && this.salary && this.year) {
      const entry: SalaryEntry = {
        name: this.selectedName,
        salary: this.salary,
        year: this.year
      };

      this.salaryService.saveEntry(entry)
        .subscribe(entries => {
          this.entries = entries;
          this.selectedName = '';
          this.salary = null;
          this.year = null;
        });
    }
  }
}

