import { Component } from '@angular/core';
import { TerminalComponent } from './terminal/terminal.component';

@Component({
  imports: [TerminalComponent],
  selector: 'app-root',
  template: '<app-terminal/>',
})
export class App {}
