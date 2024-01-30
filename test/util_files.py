import os
import shutil
from pathlib import Path
from typing import List, Tuple
import zipfile


def copy_items(source_path: Path, source_items: List[str], destination_path: Path):
    """Copy items (which may be files or directories) from source to destination."""

    # Check if source_path and destination_path are valid directories
    if not source_path.is_dir():
        raise ValueError(f"Source path is not a directory or does not exist: {source_path}")
    if not destination_path.is_dir():
        raise ValueError(
            f"Destination path is not a directory or does not exist: {destination_path}"
        )

    # Iterate over each item in the source_items list
    for item in source_items:
        source_item_path = source_path / item

        # Check if the item exists in the source directory
        if not source_item_path.exists():
            raise ValueError(f"Item '{item}' does not exist in the source directory.")

        # Define the destination item path
        destination_item_path = destination_path / item

        # Copy the item to the destination (file or directory)
        if source_item_path.is_dir():
            # Use shutil.copytree for directories, with dirs_exist_ok=True to overwrite
            shutil.copytree(source_item_path, destination_item_path, dirs_exist_ok=True)
        else:
            # Use shutil.copy2 for files, to preserve metadata
            shutil.copy2(source_item_path, destination_item_path)

def make_workspace_dir(path_workspace_dir: Path) -> Tuple[Path, Path]:
    # --------------------------
    # Create workspace
    # --------------------------
    if path_workspace_dir.exists():
        shutil.rmtree(path_workspace_dir)
    path_workspace_dir.mkdir()

    # --------------------------
    # Create standard workspace subdirectories
    # --------------------------
    path_workspace_incoming_dir = path_workspace_dir / Path('incoming')
    path_workspace_results_dir = path_workspace_dir / Path('results')

    path_workspace_incoming_dir.mkdir()
    path_workspace_results_dir.mkdir()

    return (
        path_workspace_incoming_dir,
        path_workspace_results_dir,
    )
