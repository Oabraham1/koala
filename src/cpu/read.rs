extern crate libc;
use std::time::{Duration, Instant};

use libc::{sysconf, _SC_NPROCESSORS_CONF, _SC_CLK_TCK};

struct CPUReading {
    _cpu_usage: f32,
    _time_stamp: Instant,
}

struct SystemInformation {
    _arch: str,
    _os: str,
    _chip: str,
}

pub(crate) fn cpu() {
    let num_cpus = get_num_cpus();
    let cpu_clock_speed = get_cpu_clock_speed();

    println!("Number of CPUs: {}", num_cpus);
    println!("CPU clock speed: {}", cpu_clock_speed);

    let cpu_usage = track_cpu_usage(1);

    println!("CPU Usage: {}", cpu_usage);
}

/*
    Description: This function gets the number of CPU Cores that are on the system
    Return: u16
    Parameters: None
*/
fn get_num_cpus() -> u16 {
    let num_cpu: u16 = unsafe {
        sysconf(_SC_NPROCESSORS_CONF)
            .try_into()
            .expect("Could not get number of CPUs")
    };
    num_cpu
}

/*
    Description: This function gets the number of CPU threads that are on the system
    Return: u16
    Parameters: None
*/
fn get_cpu_clock_speed() -> u16 {
    let cpu_clock_speed: u16 = unsafe {
        sysconf(_SC_CLK_TCK)
            .try_into()
            .expect("Could not get CPU clock speed")
    };
    cpu_clock_speed
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
    Description: This function reads information about the system architecture
    Return: System architectural information like CPU bit, OS, and chip information
 */
fn get_system_info() -> SystemInformation {
    SystemInformation {
        _arch: "x86".parse().unwrap(),
        _os: "macOS".parse().unwrap(),
        _chip: "Intel".parse().unwrap(),
    }
}