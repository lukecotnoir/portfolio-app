import {
  AfterViewInit, Component, ElementRef, OnInit,
  computed, inject, signal, viewChild,
} from '@angular/core';
import { CommandsService } from './commands.service';
import { L, OutLine, R, Run, TEXT } from '../model';

const PROMPT: Run[] = [R('guest@lcotnoir', 'accent'), R(':~$', 'dim')];
const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

@Component({
  selector: 'app-terminal',
  templateUrl: './terminal.component.html',
  host: { '(document:click)': 'focus()', '(document:keydown)': 'focus()' },
})
export class TerminalComponent implements OnInit, AfterViewInit {
  protected readonly cmds = inject(CommandsService);
  protected readonly output = signal<OutLine[]>([]);
  protected readonly cmd = signal('');
  protected readonly pos = signal(0);
  protected readonly clock = signal('--:--');
  protected readonly ready = signal(false);
  protected readonly mascotArt = signal('');
  protected readonly mascotBlinkArt = signal('');
  protected readonly peek = signal(false);
  protected readonly blink = signal(false);

  private input = viewChild<ElementRef<HTMLInputElement>>('kbd');
  private scroller = viewChild<ElementRef<HTMLElement>>('scroll');
  private history: string[] = [];
  private hIdx = -1;
  private busy = false;

  protected readonly before = computed(() => this.cmd().slice(0, this.pos()));
  protected readonly at = computed(() => this.cmd()[this.pos()] ?? '');
  protected readonly after = computed(() => this.cmd().slice(this.pos() + 1));
  protected readonly chips = ['help', 'work', 'skills', 'interests', 'contact', 'resume'];

  ngOnInit() {
    setInterval(() => {
      const d = new Date();
      this.clock.set(`${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`);
    }, 1000);
  }

  async ngAfterViewInit() {
    await this.cmds.load();
    this.mascotArt.set(this.cmds.content.mascot);
    this.mascotBlinkArt.set(this.cmds.content.mascotBlink ?? this.cmds.content.mascot);
    this.schedulePeek(5000 + Math.random() * 5000);
    await this.boot();
    this.ready.set(true);
    this.focus();
  }

  /* the detective in the corner: rises over the sill, blinks, ducks back.
     ~1 in 5 visits he does a double take. */
  private schedulePeek(delay: number) {
    setTimeout(() => void this.doPeek(), delay);
  }

  private async doPeek() {
    this.peek.set(true);
    const hold = 1600 + Math.random() * 1800;
    if (Math.random() < 0.55) {
      const at = 500 + Math.random() * Math.max(400, hold - 800);
      setTimeout(() => {
        this.blink.set(true);
        setTimeout(() => this.blink.set(false), 160);
      }, at);
    }
    await sleep(hold);
    this.peek.set(false);
    const next = Math.random() < 0.22
      ? 500 + Math.random() * 800
      : 14000 + Math.random() * 26000;
    this.schedulePeek(next);
  }

  /* ── output ── */

  private push(l: OutLine) {
    this.output.update((o) => [...o, l]);
    requestAnimationFrame(() => {
      const el = this.scroller()?.nativeElement;
      if (el) el.scrollTop = el.scrollHeight;
    });
  }

  private pushText(t: string, s?: Run['s'], cls?: string) {
    this.push(TEXT(t, s, cls));
  }

  private async print(lines: OutLine[], base = 24) {
    for (const l of lines) {
      this.push(l);
      const d = l.delay ?? base;
      if (d > 0) await sleep(d);
    }
  }

  private async typeLine(runs: Run[], cps = 60) {
    const text = runs.map((r) => r.t).join('');
    const line: OutLine = { runs: [R('', runs[0]?.s)] };
    this.push(line);
    for (let i = 0; i <= text.length; i++) {
      this.output.update((o) => {
        const last = o[o.length - 1];
        last.runs = [R(text.slice(0, i), runs[0]?.s)];
        return [...o];
      });
      await sleep(1000 / cps);
    }
  }

  private echo(cmd: string) {
    this.push({ runs: [...PROMPT, R(' '), R(cmd, 'bright')], cls: 'echo-line' });
  }

  /* ── boot ── */

  private async boot() {
    const c = this.cmds.content;
    const skip = sessionStorage.getItem('lc-booted');
    if (skip) {
      this.push({ runs: [R(c.banner, 'banner')] });
      this.pushText('');
      this.pushText("type 'help'. you know the drill.", 'faint');
      return;
    }
    sessionStorage.setItem('lc-booted', '1');
    for (const l of c.boot) await this.typeLine(l);
    await sleep(300);
    this.pushText('');
    this.push({ runs: [R(c.banner, 'banner')], cls: 'glitch' });
    await sleep(400);
    this.pushText('');
    await this.print(c.afterBanner.map((runs) => L(runs)), 40);
  }

  /* ── input ── */

  protected sync() {
    const el = this.input()?.nativeElement;
    if (el) this.pos.set(el.selectionStart ?? el.value.length);
  }

  protected onInput(e: Event) {
    const el = e.target as HTMLInputElement;
    this.cmd.set(el.value);
    this.sync();
  }

  protected async onKey(e: KeyboardEvent) {
    if (this.busy) {
      if (e.key !== 'Tab') e.preventDefault();
      return;
    }
    const el = e.target as HTMLInputElement;
    switch (e.key) {
      case 'Enter':
        e.preventDefault();
        await this.run(el.value);
        break;
      case 'ArrowUp':
        e.preventDefault();
        if (this.hIdx > 0) this.setCmd(this.history[--this.hIdx] ?? '');
        break;
      case 'ArrowDown':
        e.preventDefault();
        if (this.hIdx < this.history.length) this.setCmd(this.history[++this.hIdx] ?? '');
        break;
      case 'Tab': {
        e.preventDefault();
        const res = this.cmds.complete(el.value);
        if (res.value != null) this.setCmd(res.value);
        else if (res.matches) {
          this.echo(el.value);
          this.pushText(res.matches.join('   '), 'dim');
        }
        break;
      }
      case 'l':
        if (e.ctrlKey) {
          e.preventDefault();
          this.output.set([]);
        }
        break;
      case 'c':
        if (e.ctrlKey) {
          e.preventDefault();
          this.echo(el.value + '^C');
          this.setCmd('');
        }
        break;
    }
  }

  private setCmd(v: string) {
    this.cmd.set(v);
    const el = this.input()?.nativeElement;
    if (el) {
      el.value = v;
      el.selectionStart = el.selectionEnd = v.length;
    }
    this.pos.set(v.length);
  }

  protected chip(c: string) {
    if (!this.busy) this.run(c);
  }

  private async run(raw: string) {
    const trimmed = raw.trim();
    this.echo(trimmed);
    if (!trimmed) return;
    this.history.push(trimmed);
    this.hIdx = this.history.length;
    this.busy = true;
    try {
      const res = this.cmds.dispatch(trimmed);
      if (res === 'clear') this.output.set([]);
      else if (res === 'download') {
        this.pushText('fetching resume.pdf ... done', 'dim');
        const a = document.createElement('a');
        a.href = 'assets/resume.pdf';
        a.download = 'luke-cotnoir-resume.pdf';
        a.click();
        this.pushText('delivered.', 'accent');
      } else {
        await this.print(res);
      }
    } finally {
      this.busy = false;
      this.setCmd('');
      this.focus();
    }
  }

  protected focus() {
    this.input()?.nativeElement.focus({ preventScroll: true });
  }
}
