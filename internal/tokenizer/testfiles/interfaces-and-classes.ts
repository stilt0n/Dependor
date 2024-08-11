interface Foo {
  bar: string;
  baz: number;
}

export interface Extender extends Pick<Foo, 'baz'> {
  extension: 'ts';
}

export class Basic {
  constructor() {

  }

  hello() {
    console.log('hello!');
  }
}

export class Advanced extends Basic {
  constructor() {
    super();
  }

  helloAgain() {
    console.log('hello again?');
  }
}

export class Stack<T> {
  private _stack: T[]
  constructor(init?: T[]) {
    this._stack = init ?? [];
  }

  push(key: T) {
    this._stack.push(key);
    return this._stack.length;
  }

  pop() {
    return this._stack.pop();
  }

  top() {
    return this._stack[this._stack.length - 1];
  }

  size() {
    return this._stack.length;
  }
}