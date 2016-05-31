require 'fileutils'

class VerboseShell
  @@verbose = nil
  def self.verbose;     @@verbose;     end
  def self.verbose=(v); @@verbose = v; end

  def self.system_trace(*args)
    return unless @@verbose
    puts "+ "+args.map{|a| a =~ /\s/ ? '"'+a+'"' : a}.join(' ')
  end

  def self.system(*args)
    system_trace *args
    Kernel.system *args or abort "#{args[0]} failed"
  end

  def self.mv(src,dest,options={})
    system_trace *%W"mv #{src} #{dest}"
    FileUtils.mv(src, dest, options)
  end

  def self.cp(src,dest,options={})
    system_trace *%W"cp #{src} #{dest}"
    FileUtils.cp(src, dest, options)
  end

  def self.cp_r(src,dest,options={})
    system_trace *%W"cp -r #{src} #{dest}"
    FileUtils.cp_r(src, dest, options)
  end

  def self.ln_s(src,dest,options={})
    system_trace *%W"ln -s #{src} #{dest}"
    FileUtils.ln_s(src, dest, options)
  end

  def self.unlink(file)
    system_trace *%W"rm #{file}"
    File.unlink file
  end

  def self.rm_rf(file,options={})
    system_trace *%W"rm -rf #{file}"
    FileUtils.rm_rf file, options
  end

  def self.mkdir_p(file,options={})
    system_trace *%W"mkdir -p #{file}"
    FileUtils.mkdir_p file, options
  end

  def self.chown(user,group,files,options={})
    return unless user or group
    files = [files] if files.class == String
    files.each { |f| system_trace *%W"ch#{user ? 'own' : 'grp'} #{[user,group].select{|x|x}.join(':')} #{f}" }
    FileUtils.chown user, group, files, options
  end

  def self.chmod(mode, files, options = {})
    files = [files] if files.class == String
    files.each { |f| system_trace *%W"chmod #{mode} #{f}" }
    FileUtils.chmod mode, files, options
  end
end
