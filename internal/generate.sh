#!/bin/bash

antlr4 -Dlanguage=Go -package sqlparser -visitor -o ../ *.g4
