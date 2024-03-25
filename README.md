<h1 align="center">
  <br>
  Real-time task generator framework
  <br>
</h1>

<h4 align="center">A collection of task set generators for real-time systems in GO</h4>
<p align="center">
  <a href="https://github.com/porya-gohary/real-time-task-generator-framework/blob/master/LICENSE">
    <img src="https://img.shields.io/github/license/porya-gohary/real-time-task-generator-framework"
         alt="Gitter">
  </a>
    <img src="https://img.shields.io/badge/Made%20with-GO-orange">

</p>

<p align="center">
  <a href="#-required-packages">Dependencies</a> •
  <a href="#-build-instructions">Build</a> •
  <a href="#-configuration-format">Configuration Format</a> •
  <a href="#-features">Features</a> •
  <a href="#-output-format">Output format</a> •
  <a href="#-limitations">Limitations</a> •
  <a href="#-license">License</a>
</p>

## 📦 Required Packages
This program uses the following packages:

```
gopkg.in/yaml.v2
```

## 📋 Build Instructions
Considering that you already installed a recent version of GO (version 1.22.0 and higher), you can build the project using the following command:
```
go build -o generate
```
For running the program, you can use the following command:
```
./generate -config <path-to-config-file>
```
Or you can build and run the program in one step using the following command:
```
go run generate.go -config <path-to-config-file>
```

## 📝 Configuration Format
The configuration file is in YAML format.
For more information on the configuration file, please refer to the [Configuration File](examples/example-1.yaml) example.

## 🔧 Features
The framework is highly customizable and can be used to generate tasksets with different characteristics. In this section, we will discuss some of the features of the framework. 

The framework can be used to generate tasksets with the following characteristics:
- Periodic tasks
- Fork-join DAG tasks
- Random DAG tasks

To generate the periods of the tasks, the framework uses the following distribution functions:
- Uniform distribution
- Log-uniform distribution
- Uniform distribution with discrete values (which should be provided in the configuration file)
- Log-uniform distribution with discrete values (which should be provided in the configuration file)
- Automotive benchmark

Utilization of the tasks also can be generated using the following distribution functions:
- UUniFast-Discard
- RandFixedSum
- Automotive benchmark

The taskset can also be partitioned using the following partitioning algorithms:
- Best-fit
- Worst-fit
- First-fit

The framework also can unfold a generated taskset to a jobset with a specified priority assignment algorithm.
Currently, the following priority assignment algorithms are supported:
- Rate Monotonic
- Deadline Monotonic
- Earliest Deadline First

⚠️ Note: In addition to the features already listed, this framework is designed to support parallel execution. This means that multiple tasks can be run concurrently, significantly improving the performance and efficiency of the system, especially when dealing with large tasksets.

## 📄 Output Format
The generated taskset can be saved in either CSV or YAML format. 
The output format can be specified in the configuration file.


## 🚧 Limitations
- For now, the generators just support the discrete-time model and all the numbers are integer.

## 🌱 Contribution
With your feedback and conversation, you can assist me in developing this application.
- Open pull request with improvements
- Discuss feedbacks and bugs in issues

## 📜 License
Copyright © 2024 [Pourya Gohari](https://pourya-gohari.ir).

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔄 Previous Work
This project is a newer version of the [Real-time task generator](https://github.com/porya-gohary/real-time-task-generators) project. The previous project was written in Python and was not as efficient as the current project. The current project is written in GO and is designed to be more efficient, scalable and provide more options.


## 📚 References
If you are interested in the task generation algorithms, you can refer to the following papers:
* [E. Bini, G. Buttazzo, and M. Bertogna, "Measuring the Performance of Schedulability Tests," in Proceedings of the 2005 ACM Symposium on Applied Computing, 2005, pp. 1333–1337.](https://dl.acm.org/doi/abs/10.1007/s11241-005-0507-9)
* [S. Kramer, D. Ziegenbein, and A. Hamann, "Real world automotive benchmark for free"](http://rtn.ecrts.org/forum/download/WATERS15_Real_World_Automotive_Benchmark_For_Free.pdf)
* [P. Emberson, R. Stafford, and R. Davis, "Techniques for the synthesis of multiprocessor tasksets"](http://retis.sssup.it/waters2010/waters2010.pdf#page=6)
