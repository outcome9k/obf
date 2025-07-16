import os
import ast
import sys
import zlib
import base64
import string
import random
import builtins
import logging
from telegram import Update
from telegram.ext import Application, CommandHandler, MessageHandler, filters, ContextTypes

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class BlankOBFv2:
    def __init__(self, code: str, include_imports: bool=False, recursion: int=1) -> None:
        self._code = code
        self._imports = []
        self._aliases = {}
        self._valid_identifiers = [chr(x) for x in range(sys.maxunicode) if chr(x).isidentifier()]
        self.__include_imports = include_imports
        if recursion < 1:
            raise ValueError("Recursion length cannot be less than 1")
        else:
            self.__recursion = recursion

    def obfuscate(self) -> str:
        self._remove_comments_and_docstrings()
        self._save_imports()
        layers = [
            self._layer_1,
            self._layer_2,
            self._layer_3
        ] * self.__recursion
        random.shuffle(layers)
        if layers[-1] == self._layer_3:
            for index, layer in enumerate(layers):
                if layer != self._layer_3:
                    layers[index] = self._layer_3
                    layers[-1] = layer
                    break
        for layer in layers:
            layer()
        if self.__include_imports:
            self._prepend_imports()
        return self._code

    def _save_imports(self) -> None:
        def visit_node(node):
            if isinstance(node, ast.Import):
                for name in node.names:
                    self._imports.append((None, name.name))
            elif isinstance(node, ast.ImportFrom):
                module = node.module
                for name in node.names:
                    self._imports.append((module, name.name))
            for child_node in ast.iter_child_nodes(node):
                visit_node(child_node)
        tree = ast.parse(self._code)
        visit_node(tree)
        self._imports = list(set(self._imports))
        self._imports.sort(reverse=True, key=lambda x: len(x[1]) + len(x[0]) if x[0] else 0)

    def _prepend_imports(self) -> None:
        for module, submodule in self._imports:
            if module is not None:
                statement = f"from {module} import {submodule}\n"
            else:
                statement = f"import {submodule}\n"
            self._code = statement + self._code

    def _generate_random_name(self, value: str) -> str:
        if value in self._aliases.keys():
            return self._aliases.get(value)
        else:
            while True:
                name = "".join(random.choices(self._valid_identifiers, k=random.randint(10, 25)))
                if name not in self._aliases.values():
                    self._aliases[value] = name
                    return name

    def _remove_comments_and_docstrings(self) -> None:
        tree = ast.parse(self._code)
        tree.body.insert(0, ast.Expr(value=ast.Constant(":: You managed to break through BlankOBF v2; Give yourself a pat on your back! ::")))
        for index, node in enumerate(tree.body[1:]):
            if isinstance(node, ast.Expr) and isinstance(node.value, ast.Constant):
                tree.body[index] = ast.Pass()
            elif isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
                for idx, expr in enumerate(node.body):
                    if isinstance(expr, ast.Expr) and isinstance(expr.value, ast.Constant):
                        node.body[idx] = ast.Pass()
            elif isinstance(node, ast.ClassDef):
                for idx, expr in enumerate(node.body):
                    if isinstance(expr, ast.Expr) and isinstance(expr.value, ast.Constant):
                        node.body[idx] = ast.Pass()
                    elif isinstance(expr, (ast.FunctionDef, ast.AsyncFunctionDef)):
                        for i, n in enumerate(expr.body):
                            if isinstance(n, ast.Expr) and isinstance(n.value, ast.Constant):
                                expr.body[i] = ast.Pass()
        self._code = ast.unparse(tree)

    def _insert_dummy_comments(self) -> None:
        code = self._code.splitlines()
        for index in range(len(code) - 1, 0, -1):
            if random.randint(1, 10) > 3:
                spaces = 0
                comment = "#"
                for i in range(random.randint(7, 55)):
                    comment += " " + "".join(random.choices(self._valid_identifiers, k=random.randint(2, 10)))
                for i in code[index]:
                    if i != " ":
                        break
                    else:
                        spaces += 1
                code.insert(index, (" " * spaces) + comment)
        self._code = "\n".join(code)

    def _obfuscate_vars(self) -> None:
        class Transformer(ast.NodeTransformer):
            def __init__(self, outer: 'BlankOBFv2') -> None:
                self._outer = outer

            def rename(self, name: str) -> str:
                if name in dir(builtins) or name in [x[1] for x in self._outer._imports]:
                    return name
                else:
                    return self._outer._generate_random_name(name)

            def visit_Name(self, node: ast.Name) -> ast.AST:
                if node.id in dir(builtins) or node.id in [x[1] for x in self._outer._imports]:
                    node = ast.Call(
                        func=ast.Call(
                            func=ast.Name(id="getattr", ctx=ast.Load()),
                            args=[
                                ast.Call(
                                    func=ast.Name(id="__import__", ctx=ast.Load()),
                                    args=[self.visit_Constant(ast.Constant(value="builtins"))],
                                    keywords=[]
                                ),
                                self.visit_Constant(ast.Constant(value="eval"))
                            ],
                            keywords=[]
                        ),
                        args=[
                            ast.Call(
                                func=ast.Name(id="bytes", ctx=ast.Load()),
                                args=[
                                    ast.Subscript(
                                        value=ast.List(elts=[ast.Constant(value=x) for x in list(node.id.encode())][::-1], ctx=ast.Load()),
                                        slice=ast.Slice(upper=None, lower=None, step=ast.Constant(value=-1)),
                                        ctx=ast.Load()
                                    )
                                ],
                                keywords=[]
                            )
                        ],
                        keywords=[]
                    )
                    return node
                else:
                    node.id = self.rename(node.id)
                    return self.generic_visit(node)

            def visit_FunctionDef(self, node: ast.FunctionDef) -> ast.FunctionDef:
                node.name = self.rename(node.name)
                return self.generic_visit(node)

            def visit_arg(self, node: ast.arg) -> ast.arg:
                node.arg = self.rename(node.arg)
                return node

            def visit_Constant(self, node: ast.Constant) -> ast.AST:
                if isinstance(node.value, int):
                    choice = random.randint(1, 2)
                    if choice == 1:
                        num = random.randint(2**16, sys.maxsize)
                        left = node.value * num
                        right = node.value * (num - 1)
                        return ast.BinOp(left=ast.Constant(value=left), op=ast.Sub(), right=ast.Constant(value=right))
                    else:
                        num = random.randint(2**16, sys.maxsize)
                        times = random.randint(50, 500)
                        return ast.BinOp(
                            left=ast.BinOp(
                                left=ast.BinOp(
                                    left=ast.BinOp(
                                        left=ast.Constant(value=num*2),
                                        op=ast.Add(),
                                        right=ast.Constant(value=node.value*2*times),
                                    ),
                                    op=ast.FloorDiv(),
                                    right=ast.Constant(value=2),
                                ),
                                op=ast.Sub(),
                                right=ast.Constant(value=num),
                            ),
                            op=ast.Sub(),
                            right=ast.Constant(value=node.value*(times-1)),
                        )
                elif isinstance(node.value, str):
                    encoded = list(node.value.encode())[::-1]
                    return ast.Call(
                        func=ast.Attribute(
                            value=ast.Call(
                                func=ast.Name(id="bytes", ctx=ast.Load()),
                                args=[
                                    ast.Subscript(
                                        value=ast.List(elts=[ast.Constant(value=x) for x in encoded], ctx=ast.Load()),
                                        slice=ast.Slice(lower=None, upper=None, step=ast.Constant(value=-1)),
                                        ctx=ast.Load()
                                    )
                                ],
                                keywords=[]
                            ),
                            attr="decode",
                            ctx=ast.Load()
                        ),
                        args=[],
                        keywords=[]
                    )
                elif isinstance(node.value, bytes):
                    encoded = list(node.value)[::-1]
                    return ast.Call(
                        func=ast.Name(id="bytes", ctx=ast.Load()),
