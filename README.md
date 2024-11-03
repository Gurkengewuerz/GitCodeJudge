# Automated Programming Workshop Judge

A secure, scalable system for automatically testing student programming assignments in university workshops. This system integrates with Gitea for submission handling and uses Docker for secure code execution. Additionally, it seamlessly integrates with [SSHContainer](https://github.com/Gurkengewuerz/SSHContainer) to provide secure SSH access to the containers for debugging and interactive sessions.

## Features

- 🔒 **Secure Execution**: All student code runs natively in isolated Docker containers
- ⚙️ **Parallel Processing**: Handles multiple submissions simultaneously (configurable)
- ⏱️ **Real-time Feedback**: Students receive immediate test results on their commits
- 📈 **Scalable**: Handles large classes (100+ students) efficiently
- 🔐 **Privacy**: Students can't access other students' solutions (depending on the Git setup)
- 🏫 **Multiple Workshop Support**: Organize test cases by workshop and task
- 📝 **Flexible Test Cases**: Support for YAML configuration of test cases
- 🏆 **Leaderboard and Statistics**: Track student performance and display leaderboards
- 📄 **Problem PDF Exports**: Export problem statements and test cases to PDF
- 💻 **Multiple Programming Languages Support**: Supports testing code in various programming languages (currently Python, Go)
- 📅 **Time Constraints**: Set start and end dates for tasks
- 🖥️ **Interactive Development**: SSH access to development containers via SSHContainer

## Documentation

- [Quick Start Guide](docs/quick-start.md)
- [Web Handlers](docs/web-handlers.md)
- [Architecture Overview](docs/architecture.md)
- [Configuration Guide](docs/configuration.md)
- [Test Case Setup](docs/test-cases.md)
- [Instructor Guide](docs/instructor-guide.md)
- [Student Guide](docs/student-guide.md)
- [Development Guide](docs/development.md)

## License

This project is licensed under the AGPL - see the [LICENSE](LICENSE) file for details.

## Support

For support, please open an issue in the repository.
