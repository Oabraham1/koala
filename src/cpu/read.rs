use std::time::{Duration, Instant};

struct CPUReading {
    _cpu_usage: f32,
    _time_stamp: Instant,
}

struct GPUReading {
    _gpu_usage: f32,
    _time_stamp: Instant,
}

pub(crate) fn cpu() {
    let num_cpus = get_num_cpus();
    let num_gpus = get_num_gpus();
    let num_cpu_threads = get_num_cpu_threads();
    let num_gpu_threads = get_num_gpu_threads();

    println!("Number of CPUs: {}", num_cpus);
    println!("Number of GPUs: {}", num_gpus);
    println!("Number of CPU Threads: {}", num_cpu_threads);
    println!("Number of GPU Threads: {}", num_gpu_threads);

    let cpu_usage = track_cpu_usage(1);
    let gpu_usage = track_gpu_usage(1);

    println!("CPU Usage: {}", cpu_usage);
    println!("GPU Usage: {}", gpu_usage);
}

/*
    Description: This function gets the number of CPU Cores that are on the system
    Return: u16
    Parameters: None
*/
fn get_num_cpus() -> u16 {
    let num_cpu: u16 = 1;
    num_cpu
}

/*
    Description: This function gets the number of GPU cores that are on the system
    Return: u16
    Parameters: None
*/
fn get_num_gpus() -> u16 {
    let num_gpu: u16 = 1;
    num_gpu
}

/*
    Description: This function gets the number of CPU threads that are on the system
    Return: u16
    Parameters: None
*/
fn get_num_cpu_threads() -> u16 {
    let num_cpu_thread: u16 = 1;
    num_cpu_thread
}

/*
    Description: This function gets the number of GPU threads that are on the system
    Return: u16
    Parameters: None
*/
fn get_num_gpu_threads() -> u16 {
    let num_gpu_thread: u16 = 1;
    num_gpu_thread
}

/*
    Description: This function gets the percentage of CPU usage
    Return: None
    Parameters: A vector of CPUReading structs
*/
fn get_cpu_usage() -> CPUReading {
    CPUReading {
        _cpu_usage: 1.0,
        _time_stamp: Instant::now() + Duration::from_secs(3),
    }
}

/*
    Description: This function gets the percentage of GPU usage
    Return: None
    Parameters: A vector of GPUReading structs
*/
fn get_gpu_usage() -> GPUReading {
    GPUReading {
        _gpu_usage: 1.0,
        _time_stamp: Instant::now() + Duration::from_secs(3),
    }
}

/*
    Description: This function tracks the percentage of CPU usage over a specific time period
    Return: The number of usage information points that were collected
    Parameters: Time period: u32
*/
fn track_cpu_usage(time_period: u32) -> u32 {
    let start = Instant::now();
    let duration = Duration::from_secs(time_period as u64);
    let mut cpu_usage: Vec<CPUReading> = Vec::new();

    while start.elapsed() < duration {
        cpu_usage.push(get_cpu_usage());
    }
    cpu_usage.len() as u32
}

/*
    Description: This function tracks the percentage of GPU usage over a specific time period
    Return: The number of usage information points that were collected
    Parameters: Time period: u32
*/
fn track_gpu_usage(time_period: u32) -> u32 {
    let start = Instant::now();
    let duration = Duration::from_secs(time_period as u64);
    let mut gpu_usage: Vec<GPUReading> = Vec::new();

    while start.elapsed() < duration {
        gpu_usage.push(get_gpu_usage());
    }
    gpu_usage.len() as u32
}
