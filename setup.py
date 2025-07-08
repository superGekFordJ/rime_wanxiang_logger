from setuptools import setup, find_packages
import os

# Function to read the README file.
def read(fname):
    return open(os.path.join(os.path.dirname(__file__), fname), "r", encoding="utf-8").read()

setup(
    name="rime-wanxiang-logger",
    version="0.1.0",
    author="Your Name", # Replace with your name
    author_email="your_email@example.com", # Replace with your email
    description="A tool to install and manage a data logger for the Rime input method engine.",
    long_description=read('README.md'),
    long_description_content_type="text/markdown",
    url="https://github.com/your_username/rime-wanxiang-logger", # Replace with your repository URL
    license="MIT",
    packages=find_packages(),
    # This is the crucial part to ensure your .lua files are included in the package.
    # It tells setuptools that for the 'rime_logger' package, it should include
    # any files inside the 'assets' directory.
    package_data={
        'rime_logger': ['assets/*.lua'],
    },
    install_requires=[
        'pandas',
        'click',
        'questionary', # A good library for the interactive prompts
    ],
    include_package_data=True,
    # This creates the command-line script.
    # When the user installs the package, a command 'rime-logger' will be available,
    # which will execute the 'main' function in the 'rime_logger.cli' module.
    entry_points={
        'console_scripts': [
            'rime-logger=rime_logger.cli:main',
        ],
    },
    classifiers=[
        "Development Status :: 4 - Beta",
        "License :: OSI Approved :: MIT License",
        "Operating System :: MacOS :: MacOS X",
        "Operating System :: Microsoft :: Windows",
        "Operating System :: POSIX :: Linux",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.7",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Topic :: Utilities",
    ],
    python_requires='>=3.7',
)
