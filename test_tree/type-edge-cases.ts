interface EdgeCase {
  foo: string;
  bar: number;
  baz: boolean;
}

export interface UtilityInterface extends Pick<EdgeCase, 'foo' | 'baz'> {
  utils: number;
};

export type UtilityType = Omit<EdgeCase, 'foo'>;