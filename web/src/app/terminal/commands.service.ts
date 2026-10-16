import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';
import { Content, L, OutLine, R, Run, TEXT } from '../model';

type Cmd = { desc: string; run: (args: string[]) => OutLine[] | 'clear' | 'download' };

const ART_LINE = (art: string, s = 'artdark' as const): OutLine[] =>
  art.split('\n').map((t) => TEXT(t, s, 'artline'));

const FILES: Record<string, string> = {
  'about.txt': 'about',
  'experience.txt': 'work',
  'skills.txt': 'skills',
  'education.txt': 'edu',
  'contact.txt': 'contact',
  'interests.txt': 'interests',
  'resume.pdf': 'resume',
};

@Injectable({ providedIn: 'root' })
export class CommandsService {
  private http = inject(HttpClient);
  content!: Content;

  async load(): Promise<void> {
    this.content = await firstValueFrom(this.http.get<Content>('content.json'));
  }

  private section = (name: string): OutLine[] =>
    this.content.sections[name].map((runs) => L(runs));

  private table = (rows: [string, string][]): OutLine[] =>
    rows.map(([k, v]) => L([R(`  ${k.padEnd(11)}`, 'key'), R(v, 'dim')]));

  private commands(): Record<string, Cmd> {
    const c = this.content;
    return {
      help: {
        desc: 'the map',
        run: () => [
          TEXT('the essentials:', 'dim'),
          TEXT(''),
          ...this.table(c.help.essentials),
          TEXT(''),
          TEXT('for the terminal romantics:', 'dim'),
          TEXT(''),
          ...c.help.extras.map((x) => L([R('  ' + x, 'key')])),
          TEXT(''),
          TEXT(c.help.footer, 'faint'),
        ],
      },
      about: { desc: 'who this is', run: () => [...ART_LINE(c.detective), TEXT(''), ...this.section('about')] },
      whoami: { desc: 'who this is', run: () => this.commands()['about'].run([]) as OutLine[] },
      whois: { desc: 'whois luke', run: () => this.commands()['about'].run([]) as OutLine[] },
      work: { desc: 'what he ships', run: () => this.section('work') },
      experience: { desc: 'what he ships', run: () => this.section('work') },
      skills: { desc: 'the toolkit', run: () => this.section('skills') },
      edu: { desc: 'schooling', run: () => [...this.section('edu'), TEXT(''), ...this.section('skills').slice(-2)] },
      education: { desc: 'schooling', run: () => this.commands()['edu'].run([]) as OutLine[] },
      interests: { desc: 'off the clock', run: () => this.section('interests') },
      hobbies: { desc: 'off the clock', run: () => this.section('interests') },
      contact: { desc: 'get in touch', run: () => this.section('contact') },
      hire: { desc: 'the pitch', run: () => this.section('hire') },

      ls: {
        desc: 'list files',
        run: () => [
          TEXT(c.files.join('   '), 'bright'),
          TEXT('cat <file> works too.', 'faint'),
        ],
      },

      cat: {
        desc: 'cat <file>',
        run: (args) => {
          const f = args[0];
          if (!f) return [TEXT('cat: missing file operand', 'red')];
          if (f === 'secrets.txt') return [TEXT('cat: secrets.txt: permission denied', 'red')];
          const cmd = FILES[f];
          if (!cmd) return [TEXT(`cat: ${f}: no such file`, 'red')];
          return this.dispatch(cmd) as OutLine[];
        },
      },

      resume: { desc: 'download resume.pdf', run: () => 'download' },

      neofetch: {
        desc: 'system specs',
        run: () => {
          const specs = [
            ['os', 'lukeOS noir'],
            ['host', 'norfolk, va'],
            ['kernel', 'spring-boot-21'],
            ['uptime', '1.3 yrs (professional)'],
            ['shell', 'angular-cli'],
            ['wm', 'devsecops'],
            ['pkgs', 'java python ts c++'],
            ['clearance', 'ts/sci fsp'],
            ['theme', 'bone-on-black'],
            ['icons', 'none. we use text.'],
          ];
          const art = c.neofetchLogo.split('\n').filter((l) => l.length);
          const gutter = Math.max(...art.map((l) => l.length)) + 4;
          const head = 'guest@lcotnoir';
          const sep = '─'.repeat(head.length);
          const rows: OutLine[] = [];
          const n = Math.max(art.length, specs.length + 2);
          for (let i = 0; i < n; i++) {
            const a = (art[i] || '').padEnd(gutter);
            let right: Run;
            if (i === 0) right = R(head, 'accent');
            else if (i === 1) right = R(sep, 'faint');
            else {
              const s = specs[i - 2];
              right = s ? R(s[0].padEnd(10) + ' ', 'body') : R('');
            }
            const s = specs[i - 2];
            rows.push(L([R(a, 'artdark'), right, ...(s ? [R(s[1], 'dim')] : [])]));
          }
          return rows;
        },
      },

      htop: {
        desc: 'process list',
        run: () => [
          TEXT('  PID  CPU%  MEM%  PROCESS', 'dim'),
          TEXT('    1   4.2   12   spring-boot serve'),
          TEXT('  105  31.0   44   ng migrate --from react'),
          TEXT('  208   8.8    9   kubectl rollout watch'),
          TEXT('  313   0.2   61   hollow-knight --silksong --wait'),
          TEXT('  404   0.0    0   motivation.service [failed, restarted]', 'red'),
          TEXT('  700  12.1    7   guitar-amp-sim'),
          TEXT('  911   2.4    3   zillow --scroll --endless'),
          TEXT(''),
          TEXT('load average: reasonable. mostly.', 'faint'),
        ],
      },

      man: {
        desc: 'man luke',
        run: (args) => {
          if (args[0] !== 'luke') return [TEXT(`no manual entry for ${args[0] ?? ''}`.trimEnd(), 'red')];
          return [
            TEXT('LUKE(1)                    User Commands                   LUKE(1)', 'bright'),
            TEXT(''),
            TEXT('NAME'),
            TEXT('       luke - full-stack engineer, secure systems'),
            TEXT(''),
            TEXT('SYNOPSIS'),
            TEXT('       luke [--angular] [--spring] [--cleared] <mission>'),
            TEXT(''),
            TEXT('DESCRIPTION'),
            TEXT('       ships production software for places where shipping'),
            TEXT('       matters. migrates legacy frontends. runs containers'),
            TEXT('       into classified environs. documents everything.'),
            TEXT(''),
            TEXT('BUGS'),
            TEXT('       occasionally says "one more song" at 2am.'),
            TEXT(''),
            TEXT('SEE ALSO'),
            TEXT('       contact(1), hire(1), resume(1)'),
            TEXT(''),
            TEXT('LUKE(1)                    October 2026                    LUKE(1)', 'bright'),
          ];
        },
      },

      clearance: {
        desc: 'verify clearance',
        run: () => [
          L([R('> querying adjudication ...', 'dim')], undefined, 400),
          L([R('> status: ACTIVE TS/SCI — full scope polygraph', 'accent')], undefined, 400),
          L([R('> access granted. you may read this website.')], undefined, 0),
        ],
      },

      sudo: {
        desc: 'try it',
        run: (args) => {
          const joined = args.join(' ');
          if (/^hire( luke( cotnoir)?)?$/.test(joined))
            return [TEXT('permission granted. finally, a good use of sudo.', 'accent')];
          return [
            TEXT('guest is not in the sudoers file.', 'red'),
            TEXT('this incident will be reported.', 'dim'),
          ];
        },
      },

      ssh: {
        desc: 'ssh somewhere',
        run: (args) => [
          TEXT(`ssh ${args.join(' ') || 'guest@lcotnoir.dev'}`, 'dim'),
          TEXT('the real endpoint ships with the side project. watch github.', 'faint'),
        ],
      },

      exit: { desc: 'leave', run: () => [TEXT('there is no logout.', 'dim'), TEXT('only noir.', 'accent')] },
      logout: { desc: 'leave', run: () => this.commands()['exit'].run([]) as OutLine[] },
      quit: { desc: 'leave', run: () => this.commands()['exit'].run([]) as OutLine[] },

      echo: { desc: 'echo <text>', run: (args) => [TEXT(args.join(' '))] },
      pwd: { desc: 'where am i', run: () => [TEXT('/home/guest — you never left')] },
      date: { desc: 'the hour', run: () => [TEXT(new Date().toString(), 'dim')] },
      uname: { desc: 'system', run: () => [TEXT('lukeOS noir 2.6.1 tty1 x86_64')] },
      id: { desc: 'id', run: () => [TEXT('uid=1000(guest) gid=1000(visitor) groups=cool_people')] },
      ping: { desc: 'ping', run: () => [TEXT('pong. latency: zero.')] },

      clear: { desc: 'clear screen', run: () => 'clear' },
      cls: { desc: 'clear screen', run: () => 'clear' },

      vim: { desc: 'no', run: () => [TEXT('no editor installed. you get a prompt and you like it.', 'dim')] },
      emacs: { desc: 'no', run: () => this.commands()['vim'].run([]) as OutLine[] },
      nano: { desc: 'no', run: () => this.commands()['vim'].run([]) as OutLine[] },

      coffee: {
        desc: 'brew',
        run: () => [
          TEXT('        ( ('), TEXT('         ) )'), TEXT('      ........'),
          TEXT('      |      |]'), TEXT('      \\      /'), TEXT('       \\____/'),
          L([R('  black, obviously.', 'dim')], undefined, 250),
        ],
      },

      theme: {
        desc: 'change theme',
        run: (args) =>
          args[0] === 'green' || args[0] === 'matrix'
            ? [TEXT('green phosphor is overdone. staying noir.', 'red')]
            : [TEXT('one theme. it is called noir.', 'dim')],
      },

      rm: {
        desc: 'do not',
        run: (args) =>
          args.join(' ').includes('-rf') || args.join(' ').includes('/')
            ? [TEXT('rm: refused. this is a portfolio, not a prod box.', 'red')]
            : [TEXT('rm: relax.', 'red')],
      },

      games: { desc: 'easter egg', run: () => [TEXT('overwatch · wow · hollow knight.', 'dim')] },
      music: { desc: 'easter egg', run: () => [TEXT('ex-producer. guitar in the corner. always hunting the next record.', 'dim')] },
      peek: {
        desc: 'he watches',
        run: () => [...ART_LINE(c.mascot), TEXT('he sees you.', 'dim')],
      },
    };
  }

  dispatch(raw: string): OutLine[] | 'clear' | 'download' {
    const [cmd, ...args] = raw.toLowerCase().split(/\s+/);
    const entry = this.commands()[cmd];
    if (entry) return entry.run(args);
    if (cmd.startsWith('./') || cmd.startsWith('/'))
      return [TEXT(`${cmd}: permission denied`, 'red')];
    return [
      TEXT(`command not found: ${cmd}`, 'red'),
      TEXT(`type 'help'.`, 'faint'),
    ];
  }

  complete(value: string): { value?: string; matches?: string[] } {
    const parts = value.split(/\s+/);
    if (parts.length > 1 && parts[0] === 'cat') {
      const frag = parts[1] ?? '';
      const m = this.content.files.filter((f) => f.startsWith(frag));
      return m.length === 1 ? { value: `cat ${m[0]}` } : { matches: m.length ? m : undefined };
    }
    const frag = parts[0] ?? '';
    const m = Object.keys(this.commands()).filter((k) => k.startsWith(frag));
    return m.length === 1 ? { value: m[0] } : { matches: frag && m.length ? m : undefined };
  }

  fileFor(name: string): string | undefined {
    return FILES[name];
  }
}
