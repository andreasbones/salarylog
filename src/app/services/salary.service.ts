import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable } from 'rxjs';

export interface SalaryEntry {
  name: string;
  salary: number;
  year: number;
}

@Injectable({
  providedIn: 'root'
})
export class SalaryService {
  private apiUrl = 'http://localhost:8080';
  
  private namesSubject = new BehaviorSubject<string[]>([]);
  private entriesSubject = new BehaviorSubject<SalaryEntry[]>([]);

  names$ = this.namesSubject.asObservable();
  entries$ = this.entriesSubject.asObservable();

  constructor(private http: HttpClient) {
    this.loadInitialData();
  }

  private loadInitialData() {
    this.http.get<{names: string[], entries: SalaryEntry[]}>(`${this.apiUrl}/data`).subscribe(
      data => {
        this.namesSubject.next(data.names);
        this.entriesSubject.next(data.entries);
      }
    );
  }

  addName(name: string): Observable<string[]> {
    return this.http.post<string[]>(`${this.apiUrl}/names`, { name });
  }

  saveEntry(entry: SalaryEntry): Observable<SalaryEntry[]> {
    return this.http.post<SalaryEntry[]>(`${this.apiUrl}/entries`, entry);
  }
}

