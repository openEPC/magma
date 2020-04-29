#!/usr/bin/env python3
# @generated AUTOGENERATED file. Do not Change!

from dataclasses import dataclass
from datetime import datetime
from functools import partial
from gql.gql.datetime_utils import DATETIME_FIELD
from numbers import Number
from typing import Any, Callable, List, Mapping, Optional

from dataclasses_json import DataClassJsonMixin

from gql.gql.enum_utils import enum_field
from .file_type_enum import FileType

@dataclass
class FileInput(DataClassJsonMixin):
    fileName: str
    storeKey: str
    id: Optional[str] = None
    sizeInBytes: Optional[int] = None
    modificationTime: Optional[int] = None
    uploadTime: Optional[int] = None
    fileType: Optional[FileType] = None
    mimeType: Optional[str] = None
