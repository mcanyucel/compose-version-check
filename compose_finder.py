#!/usr/bin/env python3
import os
import sys
import shutil
from pathlib import Path
import yaml

class ComposeFileFinder:
    def __init__(self, source_dir, project_dir):
        self.source_dir = Path(source_dir).resolve()
        self.project_dir = Path(project_dir).resolve()
        self.containers_dir = self.project_dir / "containers"
        self.config_file = self.project_dir / "config.yaml"
        
    def find_compose_files(self):
        """Find all docker-compose files in the source directory."""
        compose_files = []
        for root, _, files in os.walk(self.source_dir):
            for file in files:
                if file in ["docker-compose.yml", "docker-compose.yaml"]:
                    compose_files.append(Path(root) / file)
        return compose_files
    
    def copy_compose_files(self, compose_files):
        """Copy compose files to the project directory."""
        file_mappings = []
        self.containers_dir.mkdir(exist_ok=True)
        
        for source_file in compose_files:
            # Get relative container name from parent directory
            container_name = source_file.parent.name
            target_dir = self.containers_dir / container_name
            target_dir.mkdir(exist_ok=True)
            
            # Copy the file
            target_file = target_dir / source_file.name
            shutil.copy2(source_file, target_file)
            
            # Store the mapping
            file_mappings.append({
                'source': str(source_file),
                'local_path': f"containers/{container_name}/{source_file.name}",
                'container_name': container_name
            })
            
        return file_mappings
    
    def update_config(self, file_mappings):
        """Update the config.yaml file with new mappings."""
        config = {
            'files': [],
            'notifications': {
                'type': 'debug',  # Default to debug mode
                'debug_file': 'notifications'
            }
        }
        
        # Read existing config if it exists
        if self.config_file.exists():
            with open(self.config_file) as f:
                config = yaml.safe_load(f) or config
        
        # Update files section
        existing_paths = {f['local_path'] for f in config['files'] if 'local_path' in f}
        
        for mapping in file_mappings:
            if mapping['local_path'] not in existing_paths:
                config['files'].append({
                    'local_path': mapping['local_path'],
                    'source_url': '# TODO: Add source URL for ' + mapping['container_name']
                })
        
        # Write updated config
        with open(self.config_file, 'w') as f:
            yaml.dump(config, f, default_flow_style=False)

def main():
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} <source_directory>")
        sys.exit(1)
    
    source_dir = sys.argv[1]
    project_dir = Path.cwd()
    
    finder = ComposeFileFinder(source_dir, project_dir)
    
    print(f"Scanning {source_dir} for docker-compose files...")
    compose_files = finder.find_compose_files()
    
    if not compose_files:
        print("No docker-compose files found!")
        sys.exit(1)
    
    print(f"Found {len(compose_files)} docker-compose files")
    file_mappings = finder.copy_compose_files(compose_files)
    finder.update_config(file_mappings)
    
    print("\nProcessed files:")
    for mapping in file_mappings:
        print(f"\n{mapping['container_name']}:")
        print(f"  Source: {mapping['source']}")
        print(f"  Local: {mapping['local_path']}")
    
    print("\nConfiguration updated. Please:")
    print("1. Review the config.yaml file")
    print("2. Add source URLs for each compose file")
    print("3. Configure your preferred notification method")

if __name__ == "__main__":
    main()
