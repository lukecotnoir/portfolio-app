export type Style =
  | 'body' | 'dim' | 'faint' | 'accent' | 'bright'
  | 'red' | 'key' | 'art' | 'artdark' | 'banner';

export interface Run {
  t: string;
  s?: Style;
  href?: string;
}

export type Line = Run[];

export interface OutLine {
  runs: Run[];
  cls?: string;
  delay?: number;
}

export interface Content {
  banner: string;
  detective: string;
  neofetchLogo: string;
  mascot: string;
  mascotLeft?: string;
  mascotRight?: string;
  mascotBlink?: string;
  boot: Line[];
  afterBanner: Line[];
  files: string[];
  sections: Record<string, Line[]>;
  help: {
    essentials: [string, string][];
    extras: string[];
    footer: string;
  };
}

export const L = (runs: Run[], cls?: string, delay?: number): OutLine => ({ runs, cls, delay });
export const R = (t: string, s?: Style, href?: string): Run => ({ t, s, href });
export const TEXT = (t: string, s?: Style, cls?: string): OutLine => L([R(t, s)], cls);
