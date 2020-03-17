# -*- coding: utf-8 -*-
import os

BLACK_LIST_FOLDERS = [
                    ".git", 
                    "tests/jenkins"
                     ]


def is_gitroot_file(filename):
    return not bool(os.path.split(filename)[0])

def in_blacklist_folder(filename):
    return any(filename.startswith(folder) for folder in BLACK_LIST_FOLDERS)
