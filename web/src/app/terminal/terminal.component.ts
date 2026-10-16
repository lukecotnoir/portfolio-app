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
  protected readonly mascotFrames = signal<string[]>([]);
  protected readonly frameIdx = signal(0);
  protected readonly peek = signal(false);

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
    const c = this.cmds.content;
    this.mascotFrames.set([c.mascot, c.mascotLeft ?? c.mascot, c.mascotRight ?? c.mascot, c.mascotBlink ?? c.mascot]);
    this.schedulePeek(5000 + Math.random() * 5000);
    this.startGlitchLoop();
    await this.boot();
    this.ready.set(true);
    this.focus();
  }

  /* the detective in the corner: rises over the sill, looks left, looks
     right, and ducks back out. ~1 in 5 visits he does a double take. */
  private schedulePeek(delay: number) {
    setTimeout(() => void this.doPeek(), delay);
  }

  private async doPeek() {
    const f = this.frameIdx;
    this.peek.set(true);
    f.set(0);
    await sleep(650); f.set(1);                 // eyes left
    await sleep(550); f.set(2);                 // eyes right
    await sleep(550); f.set(0);                 // back to you
    if (Math.random() < 0.35) {
      await sleep(200); f.set(3);               // blink
      await sleep(170); f.set(0);
      await sleep(280);
    } else {
      await sleep(400);
    }
    this.peek.set(false);
    const next = Math.random() < 0.22
      ? 550 + Math.random() * 800
      : 14000 + Math.random() * 26000;
    this.schedulePeek(next);
  }

  /* signal noise: every few seconds a few glyphs in whatever art is on
     screen briefly corrupt, then heal. */
  private startGlitchLoop() {
    const GLYPHS = " .'`^\",:;Il!i><~+_-?][}{1)(|\\/tfjrxnuvczXYUJCLQ0OZmwqpdbkhao*#MW&8%B$@#%&";
    const tick = () => {
      const idxs: number[] = [];
      this.output().forEach((l, i) => { if (l.cls === 'artline') idxs.push(i); });
      const bursts = Math.random() < 0.25 && idxs.length > 2 ? 3 : 1;
      for (let b = 0; b < bursts && idxs.length; b++) {
        const i = idxs[Math.floor(Math.random() * idxs.length)];
        const line = this.output()[i];
        const run = line.runs[0];
        const chars = run.t.split('');
        const n = 2 + Math.floor(Math.random() * 6);
        for (let k = 0; k < n; k++) {
          const p = Math.floor(Math.random() * chars.length);
          if (chars[p] !== ' ') {
            chars[p] = Math.random() < 0.2 ? ' ' : GLYPHS[Math.floor(Math.random() * GLYPHS.length)];
          }
        }
        const corrupt = chars.join('');
        this.output.update((o) => {
          const j = o.indexOf(line);
          if (j < 0) return o;
          const c = [...o];
          c[j] = { ...line, runs: [{ ...run, t: corrupt }] };
          return c;
        });
        setTimeout(() => {
          this.output.update((o) => {
            const j = o.indexOf(line);
            if (j < 0) return o;
            const c = [...o];
            c[j] = { ...line, runs: [run] };
            return c;
          });
        }, 80 + Math.random() * 140);
      }
      setTimeout(tick, 2200 + Math.random() * 5500);
    };
    setTimeout(tick, 4000);
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
